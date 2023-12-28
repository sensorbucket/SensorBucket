package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
	// Setup API keys service
	db, err := createDB()
	if err != nil {
		panic(err)
	}
	tenantSTore := tenantsinfra.NewTenantsStorePSQL(db)
	s := tenants.NewTenantService(tenantSTore)
	tenantHttp := tenantstransports.NewTenantsHTTP(s, "localhost")
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      tenantHttp,
	}
	srv.ListenAndServe()
	return
	// apiKeyStore := tenantsinfra.NewAPIKeyStorePSQL(db)
	// apiKeySvc := apikeys.NewAPIKeyService(&tmock{}, apiKeyStore)
	// apiKeyHttp := tenantstransports.NewAPIKeysHTTP(apiKeySvc, "localhost")
	// srv := &http.Server{
	// 	Addr:         ":3010",
	// 	WriteTimeout: 5 * time.Second,
	// 	ReadTimeout:  5 * time.Second,
	// 	Handler:      apiKeyHttp,
	// }
	// srv.ListenAndServe()

	log.Printf("[Info] Running Tenants API on %s\n", HTTP_ADDR)
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

type tmock struct{}

func (t *tmock) GetTenantById(id int64) (apikeys.Tenant, error) {
	return apikeys.Tenant{
		ID:    5,
		State: apikeys.Active,
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
