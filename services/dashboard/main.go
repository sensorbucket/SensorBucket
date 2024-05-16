package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/dashboard/routes"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

func main() {
	if err := Run(); err != nil {
		panic(fmt.Sprintf("error: %v\n", err))
	}
}

var (
	HTTP_ADDR       = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE       = env.Could("HTTP_BASE", "")
	AUTH_JWKS_URL   = env.Could("AUTH_JWKS_URL", "http://oathkeeper:4456/.well-known/jwks.json")
	EP_CORE         = env.Must("EP_CORE")
	EP_WORKERS      = env.Must("EP_WORKERS")
	EP_TRACING      = env.Must("EP_TRACING")
	EP_MEASUREMENTS = env.Must("EP_MEASUREMENTS")
)

//go:embed static/*
var staticFS embed.FS

func Run() error {
	errC := make(chan error, 1)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	router := chi.NewRouter()
	jwks := auth.NewJWKSHttpClient(AUTH_JWKS_URL)
	router.Use(
		middleware.Logger,
		auth.ForwardRequestAuthentication(),
		auth.Authenticate(jwks),
		// auth.Protect(),
	)

	var baseURL *url.URL
	if HTTP_BASE != "" {
		baseURL, _ = url.Parse(HTTP_BASE)
		views.SetBase(baseURL)
	}

	// Serve static files
	fileServer := http.FileServer(http.FS(staticFS))
	router.Handle("/static/*", fileServer)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		u := "/overview"
		if baseURL != nil {
			u = baseURL.JoinPath("overview").String()
		}
		http.Redirect(w, r, u, http.StatusFound)
	})
	router.Mount("/overview", routes.CreateOverviewPageHandler(
		createAPIClient(EP_CORE),
		createAPIClient(EP_MEASUREMENTS),
	))
	router.Mount("/ingress", routes.CreateIngressPageHandler(
		createAPIClient(EP_CORE),
		createAPIClient(EP_TRACING),
		createAPIClient(EP_WORKERS),
	))
	router.Mount("/workers", routes.CreateWorkerPageHandler(
		createAPIClient(EP_WORKERS),
	))
	router.Mount("/pipelines", routes.CreatePipelinePageHandler(
		createAPIClient(EP_WORKERS),
		createAPIClient(EP_CORE),
	))
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
	var err error
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

func createAPIClient(baseurl string) *api.APIClient {
	sbURL, err := url.Parse(baseurl)
	if err != nil {
		log.Fatalf("could not parse APIClient url: %s\n", err)
	}
	cfg := api.NewConfiguration()
	cfg.Scheme = sbURL.Scheme
	cfg.Host = sbURL.Host
	return api.NewAPIClient(cfg)
}
