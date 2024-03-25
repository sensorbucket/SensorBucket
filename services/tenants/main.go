package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/ory/nosurf"

	"sensorbucket.nl/sensorbucket/internal/env"
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
	SB_API           = env.Must("SB_API")
	DB_DSN           = env.Must("DB_DSN")
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
}

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	errC := make(chan error, 1)

	// Setup API keys service
	db, err := createDB()
	if err != nil {
		panic(err)
	}

	stopAPI, err := runAPI(errC, db)
	if err != nil {
		return fmt.Errorf("could not setup API server: %w", err)
	}
	stopWebUI, err := runWebUI(errC, db)
	if err != nil {
		return fmt.Errorf("could not setup WebUI server: %w", err)
	}

	select {
	case err = <-errC:
	case <-ctx.Done():
	}

	ctxTO, cancelTO := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTO()

	stopAPI(ctxTO)
	stopWebUI(ctxTO)

	return err
}

var noopCleanup = func(ctx context.Context) {}

func runAPI(errC chan<- error, db *sqlx.DB) (func(context.Context), error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Setup Tenants service
	tenantStore := tenantsinfra.NewTenantsStorePSQL(db)
	kratosAdmin := tenantsinfra.NewKratosUserValidator(KRATOS_ADMIN_API)
	tenantSvc := tenants.NewTenantService(tenantStore, kratosAdmin)
	_ = tenantstransports.NewTenantsHTTP(r, tenantSvc, HTTP_API_BASE)

	// Setup API keys service
	apiKeyStore := tenantsinfra.NewAPIKeyStorePSQL(db)
	apiKeySvc := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)
	_ = tenantstransports.NewAPIKeysHTTP(r, apiKeySvc, HTTP_API_BASE)

	// Run the HTTP Server
	srv := &http.Server{
		Addr:         HTTP_API_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      r,
	}

	go func() {
		log.Printf("[Info] Running Tenants API on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			errC <- err
		}
	}()

	return func(shutdownCtx context.Context) {
		if err := srv.Shutdown(shutdownCtx); !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Printf("API HTTP Server error during shutdown: %v\n", err)
		}
	}, nil
}

func runWebUI(errC chan<- error, db *sqlx.DB) (func(context.Context), error) {
	// Setup Tenants service
	tenantStore := tenantsinfra.NewTenantsStorePSQL(db)
	kratosAdmin := tenantsinfra.NewKratosUserValidator(KRATOS_ADMIN_API)
	tenantSvc := tenants.NewTenantService(tenantStore, kratosAdmin)
	userPreferences := sessions.NewUserPreferenceService(tenantStore)
	apiKeyStore := tenantsinfra.NewAPIKeyStorePSQL(db)
	apiKeySvc := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)

	ui, err := webui.New(HTTP_WEBUI_BASE, SB_API, tenantSvc, apiKeySvc, userPreferences)
	if err != nil {
		errC <- err
		return noopCleanup, nil
	}

	httpHandler := nosurf.New(ui)
	httpHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("nosurf.Reason(r): %v\n", nosurf.Reason(r))
		w.Header().Add("HX-Trigger", `{"error":"CSRF token was invalid"}`)
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

	return func(shutdownCtx context.Context) {
		if err := srv.Shutdown(shutdownCtx); !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Printf("WebUI HTTP Server error during shutdown: %v\n", err)
		}
	}, nil
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
