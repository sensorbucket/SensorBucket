package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	tenantsinfra "sensorbucket.nl/sensorbucket/services/tenants/infrastructure"
	"sensorbucket.nl/sensorbucket/services/tenants/migrations"
	tenantstransports "sensorbucket.nl/sensorbucket/services/tenants/transports"
)

func main() {
	// Setup API keys service
	db, err := createDB()
	if err != nil {
		panic(err)
	}
	apiKeyStore := tenantsinfra.NewAPIKeyStorePSQL(db)
	apiKeySvc := apikeys.NewAPIKeyService(&tmock{}, apiKeyStore)
	apiKeyHttp := tenantstransports.NewAPIKeysHTTP(apiKeySvc, "localhost")
	srv := &http.Server{
		Addr:         ":3010",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      apiKeyHttp,
	}
	srv.ListenAndServe()

}

type tmock struct{}

func (t *tmock) GetTenantById(id int64) (apikeys.Tenant, error) {
	return apikeys.Tenant{
		ID:    1,
		State: apikeys.Active,
	}, nil
}

func createDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", "postgresql://sensorbucket:sensorbucket@localhost:5432/tenants?sslmode=disable")
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
