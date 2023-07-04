package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infrastructure"
	devicetransport "sensorbucket.nl/sensorbucket/services/core/devices/transport"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	measurementsinfra "sensorbucket.nl/sensorbucket/services/core/measurements/infra"
	measurementtransport "sensorbucket.nl/sensorbucket/services/core/measurements/transport"
	"sensorbucket.nl/sensorbucket/services/core/migrations"
)

var (
	DB_DSN           = env.Must("DB_DSN")
	HTTP_ADDR        = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE        = env.Could("HTTP_BASE", "http://localhost:3000/api")
	SYS_ARCHIVE_TIME = env.Could("SYS_ARCHIVE_TIME", "30")
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
	}
}

func Run() error {
	db, err := createDB()
	if err != nil {
		return fmt.Errorf("could not create database connection: %w", err)
	}

	devicestore := deviceinfra.NewPSQLStore(db)
	deviceservice := devices.New(devicestore)
	deviceshttp := devicetransport.NewHTTPTransport(deviceservice, HTTP_BASE)

	sysArchiveTime, err := strconv.Atoi(SYS_ARCHIVE_TIME)
	if err != nil {
		return fmt.Errorf("could not convert SYS_ARCHIVE_TIME to integer: %w", err)
	}
	measurementstore := measurementsinfra.NewPSQL(db)
	measurementservice := measurements.New(measurementstore, sysArchiveTime)
	measurementhttp := measurementtransport.NewHTTP(measurementservice, HTTP_BASE)

	r := chi.NewRouter()
	r.Mount("/devices", deviceshttp)
	r.Mount("/measurements", measurementhttp)
	httpsrv := createHTTPServer(r)

	// TODO: make better
	return httpsrv.ListenAndServe()
}

func createHTTPServer(h http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      cors.AllowAll().Handler(h),
	}
	return srv
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
