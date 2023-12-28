package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	tenantsinfra "sensorbucket.nl/sensorbucket/services/tenants/infrastructure"
	"sensorbucket.nl/sensorbucket/services/tenants/migrations"
	tenantstransports "sensorbucket.nl/sensorbucket/services/tenants/transports"
)

var (
	HTTP_ADDR = env.Could("HTTP_ADDR", ":3010")
	HTTP_BASE = env.Could("HTTP_BASE", "http://localhost:3010/api")
	DB_DSN    = env.Must("DB_DSN")
)

func main() {
	// Setup API keys service
	db, err := createDB()
	if err != nil {
		panic(err)
	}
	apiKeyStore := tenantsinfra.NewAPIKeyStorePSQL(db)
	apiKeySvc := apikeys.NewAPIKeyService(&tmock{}, apiKeyStore)
	apiKeyHttp := tenantstransports.NewAPIKeysHTTP(apiKeySvc, HTTP_BASE)
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      apiKeyHttp,
	}
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
