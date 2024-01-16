package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/dashboard/routes"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

func main() {
	if err := Run(); err != nil {
		panic(fmt.Sprintf("error: %v\n", err))
	}
}

var (
	startTS   = time.Now()
	HTTP_ADDR = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE = env.Could("HTTP_BASE", "")
	SB_API    = env.Must("SB_API")
)

//go:embed static/*
var staticFS embed.FS

func Run() error {
	errC := make(chan error, 1)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	var baseURL *url.URL
	if HTTP_BASE != "" {
		baseURL, _ = url.Parse(HTTP_BASE)
		views.SetBase(baseURL)
	}

	// Middleware to pass on basic auth to the client api
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if ok {
				r = r.WithContext(context.WithValue(
					r.Context(), api.ContextBasicAuth, api.BasicAuth{
						UserName: user,
						Password: pass,
					}))
			}
			next.ServeHTTP(w, r)
		})
	})

	// Serve static files
	fileServer := http.FileServer(http.FS(staticFS))
	router.Handle("/static/*", fileServer)

	sbURL, err := url.Parse(SB_API)
	if err != nil {
		return fmt.Errorf("could not parse SB_API url: %w", err)
	}
	cfg := api.NewConfiguration()
	cfg.Scheme = sbURL.Scheme
	cfg.Host = sbURL.Host
	apiClient := api.NewAPIClient(cfg)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		u := "/overview"
		if baseURL != nil {
			u = baseURL.JoinPath("overview").String()
		}
		http.Redirect(w, r, u, http.StatusFound)
	})
	router.Mount("/overview", routes.CreateOverviewPageHandler(apiClient))
	router.Mount("/ingress", routes.CreateIngressPageHandler(apiClient))
	router.Mount("/workers", routes.CreateWorkerPageHandler(apiClient))
	router.Mount("/pipelines", routes.CreatePipelinePageHandler(apiClient))
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
	}()
	fmt.Printf("HTTP Server listening on: %s\n", srv.Addr)

	// Wait for fatal error or interrupt signal
	select {
	case <-ctx.Done():
	case err = <-errC:
		cancel()
	}

	ctxTO, cancelTO := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTO()

	if err := srv.Shutdown(ctxTO); err != nil {
		fmt.Printf("could not gracefully shutdown http server: %s\n", err)
	}

	return err
}
