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
	"github.com/ory/nosurf"

	"sensorbucket.nl/sensorbucket/internal/buildinfo"
	"sensorbucket.nl/sensorbucket/internal/cleanupper"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/healthchecker"
	"sensorbucket.nl/sensorbucket/pkg/layout"
	"sensorbucket.nl/sensorbucket/services/dashboard/routes"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

func main() {
	buildinfo.Print()
	cleanup := cleanupper.Create()
	defer func() {
		if err := cleanup.Execute(5 * time.Second); err != nil {
			log.Printf("[Warn] Cleanup error(s) occured: %s\n", err)
		}
	}()
	if err := Run(cleanup); err != nil {
		log.Fatalf("error: %s\n", err)
	}
}

var (
	HTTP_ADDR     = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE     = env.Could("HTTP_BASE", "")
	AUTH_JWKS_URL = env.Could("AUTH_JWKS_URL", "http://oathkeeper:4456/.well-known/jwks.json")
	EP_CORE       = env.Must("EP_CORE")
	EP_WORKERS    = env.Must("EP_WORKERS")
	EP_TRACING    = env.Must("EP_TRACING")
)

//go:embed static/*
var staticFS embed.FS

func Run(cleanup cleanupper.Cleanupper) error {
	errC := make(chan error, 1)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	stopProfiler, err := web.RunProfiler()
	if err != nil {
		fmt.Printf("could not setup profiler server: %s\n", err)
	}
	cleanup.Add(stopProfiler)

	router := chi.NewRouter()
	jwks := auth.NewJWKSHttpClient(AUTH_JWKS_URL)
	router.Use(
		middleware.Logger,
		auth.ForwardRequestAuthentication(),
		auth.Authenticate(jwks),
		auth.Protect(),
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
	csrfWrappedHandler := nosurf.New(router)
	csrfWrappedHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("nosurf.Reason(r): %v\n", nosurf.Reason(r))
		layout.WithSnackbarError(w, "CSRF Token was invalid, try reloading the page")
		//nolint
		w.Write([]byte("A CSRF error occured. Reload the previous page and try again"))
	}))
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      csrfWrappedHandler,
	}
	cleanup.Add(srv.Shutdown)

	shutdownHealthServer := healthchecker.Create().WithEnv().Start(ctx)
	cleanup.Add(shutdownHealthServer)

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

	return err
}

func createAPIClient(baseurl string) *api.APIClient {
	cfg := api.NewConfiguration()
	cfg.Servers = api.ServerConfigurations{
		{
			URL: baseurl,
		},
	}
	return api.NewAPIClient(cfg)
}
