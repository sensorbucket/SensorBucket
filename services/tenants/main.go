package main

import (
	"fmt"
	"log"
	"net/http"
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
)

var (
	HTTP_ADDR = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE = env.Could("HTTP_BASE", "http://localhost:3000/api")
	DB_DSN    = env.Must("DB_DSN")
)

func main() {
	// Setup DB connection
	db, err := createDB()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Setup Tenants service
	tenantStore := tenantsinfra.NewTenantsStorePSQL(db)
	tenantSvc := tenants.NewTenantService(tenantStore)
	_ = tenantstransports.NewTenantsHTTP(r, tenantSvc, HTTP_BASE)

	// Setup API keys service
	apiKeyStore := tenantsinfra.NewAPIKeyStorePSQL(db)
	apiKeySvc := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)
	_ = tenantstransports.NewAPIKeysHTTP(r, apiKeySvc, HTTP_BASE)

	// Run the HTTP Server
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      r,
	}

	log.Printf("[Info] Running Tenants API on %s\n", HTTP_ADDR)
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
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
