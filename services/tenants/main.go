package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/cors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/ory/nosurf"

	"sensorbucket.nl/sensorbucket/internal/buildinfo"
	"sensorbucket.nl/sensorbucket/internal/cleanupper"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/healthchecker"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	tenantsinfra "sensorbucket.nl/sensorbucket/services/tenants/infrastructure"
	"sensorbucket.nl/sensorbucket/services/tenants/migrations"
	"sensorbucket.nl/sensorbucket/services/tenants/sessions"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
	tenantstransports "sensorbucket.nl/sensorbucket/services/tenants/transports"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui"
)

var (
	HTTP_API_ADDR    = env.Could("HTTP_ADDR", ":3000")
	HTTP_API_BASE    = env.Could("HTTP_BASE", "http://localhost:3000/api")
	HTTP_WEBUI_ADDR  = env.Could("HTTP_WEBUI_ADDR", ":3001")
	HTTP_WEBUI_BASE  = env.Could("HTTP_WEBUI_BASE", "http://localhost:3000/auth")
	KRATOS_ADMIN_API = env.Could("KRATOS_ADMIN_API", "http://kratos:4434/")
	AUTH_JWKS_URL    = env.Could("AUTH_JWKS_URL", "http://oathkeeper:4456/.well-known/jwks.json")
	DB_DSN           = env.Must("DB_DSN")
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
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
}

func Run(cleanup cleanupper.Cleanupper) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	errC := make(chan error, 1)

	// Setup API keys service
	db, err := createDB()
	if err != nil {
		panic(err)
	}

	stopProfiler, err := web.RunProfiler()
	if err != nil {
		fmt.Printf("could not setup profiler server: %s\n", err)
	}
	cleanup.Add(stopProfiler)
	stopAPI, err := runAPI(errC, db)
	if err != nil {
		return fmt.Errorf("could not setup API server: %w", err)
	}
	cleanup.Add(stopAPI)
	stopWebUI, err := runWebUI(errC, db)
	if err != nil {
		return fmt.Errorf("could not setup WebUI server: %w", err)
	}
	cleanup.Add(stopWebUI)

	shutdownHealth := healthchecker.Create().WithEnv().Start(ctx)
	cleanup.Add(shutdownHealth)

	select {
	case err = <-errC:
	case <-ctx.Done():
	}

	return err
}

var noopCleanup = func(ctx context.Context) error { return nil }

func runAPI(errC chan<- error, db *sqlx.DB) (func(context.Context) error, error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	jwks := auth.NewJWKSHttpClient(AUTH_JWKS_URL)
	authMW := auth.Authenticate(jwks)

	// Setup Tenants service
	tenantStore := tenantsinfra.NewTenantsStorePSQL(db)
	kratosAdmin := tenantsinfra.NewKratosUserValidator(KRATOS_ADMIN_API)
	tenantSVC := tenants.NewTenantService(tenantStore, kratosAdmin)
	_ = tenantstransports.NewTenantsHTTP(r.With(authMW), tenantSVC, HTTP_API_BASE)

	// Setup API keys service
	apiKeyStore := tenantsinfra.NewAPIKeyStorePSQL(db)
	apiKeySVC := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)
	_ = tenantstransports.NewAPIKeysHTTP(r.With(authMW), apiKeySVC, HTTP_API_BASE)

	// Setup oathkeeper endpoint
	userPreferences := sessions.NewUserPreferenceService(tenantStore, tenantStore)
	oathkeeperTransport := tenantstransports.NewOathkeeperEndpoint(userPreferences, tenantSVC)
	r.Mount("/oathkeeper", oathkeeperTransport)

	// Run the HTTP Server
	srv := &http.Server{
		Addr:         HTTP_API_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      cors.AllowAll().Handler(r),
	}

	go func() {
		log.Printf("[Info] Running Tenants API on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			errC <- err
		}
	}()

	return srv.Shutdown, nil
}

func runWebUI(errC chan<- error, db *sqlx.DB) (func(context.Context) error, error) {
	// Setup Tenants service
	tenantStore := tenantsinfra.NewTenantsStorePSQL(db)
	kratosAdmin := tenantsinfra.NewKratosUserValidator(KRATOS_ADMIN_API)
	tenantSvc := tenants.NewTenantService(tenantStore, kratosAdmin)
	userPreferences := sessions.NewUserPreferenceService(tenantStore, tenantStore)
	apiKeyStore := tenantsinfra.NewAPIKeyStorePSQL(db)
	apiKeySvc := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)

	ui, err := webui.New(
		HTTP_WEBUI_BASE,
		AUTH_JWKS_URL,
		tenantSvc,
		apiKeySvc,
		userPreferences,
	)
	if err != nil {
		errC <- err
		return noopCleanup, nil
	}

	httpHandler := nosurf.New(ui)
	httpHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("nosurf.Reason(r): %v\n", nosurf.Reason(r))
		w.Header().Add("HX-Trigger", `{"error":"CSRF token was invalid"}`)
		//nolint
		w.Write([]byte("A CSRF error occured. Reload the previous page and try again"))
	}))
	srv := &http.Server{
		Addr:         HTTP_WEBUI_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      httpHandler,
	}

	go func() {
		log.Printf("[Info] Running Tenants WebUI on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			errC <- err
		}
	}()

	return srv.Shutdown, nil
}

func createDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", DB_DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}
	return db, nil
}
