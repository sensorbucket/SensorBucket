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
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	tenantsinfra "sensorbucket.nl/sensorbucket/services/tenants/infrastructure"
	"sensorbucket.nl/sensorbucket/services/tenants/migrations"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
	tenantstransports "sensorbucket.nl/sensorbucket/services/tenants/transports"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui"
)

var (
	HTTP_API_ADDR   = env.Could("HTTP_ADDR", ":3000")
	HTTP_API_BASE   = env.Could("HTTP_BASE", "http://localhost:3000/api")
	HTTP_WEBUI_ADDR = env.Could("HTTP_WEBUI_ADDR", ":3001")
	HTTP_WEBUI_BASE = env.Could("HTTP_WEBUI_BASE", "http://localhost:3000/auth")
	DB_DSN          = env.Must("DB_DSN")
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

	stopAPI := runAPI(errC, db)
	stopWebUI := runWebUI(errC)

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

func runAPI(errC chan<- error, db *sqlx.DB) func(context.Context) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Setup Tenants service
	tenantStore := tenantsinfra.NewTenantsStorePSQL(db)
	tenantSvc := tenants.NewTenantService(tenantStore)
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
	}
}

func runWebUI(errC chan<- error) func(context.Context) {
	ui := webui.New(HTTP_WEBUI_BASE)
	srv := &http.Server{
		Addr:         HTTP_WEBUI_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      ui,
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
	}
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
