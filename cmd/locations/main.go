package main

import (
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"sensorbucket.nl/internal/locations/http"
	"sensorbucket.nl/internal/locations/store"
)

const (
	LOCATION_SVC_HTTP_HOST     = "LOCATION_SVC_HTTP_HOST"     // :8080
	LOCATION_SVC_WORKER_DB_DSN = "LOCATION_SVC_WORKER_DB_DSN" // postgresql://root:root@localhost:5432/todos?sslmode=disable
)

func main() {
	db, err := sqlx.Open("pgx", os.Getenv(LOCATION_SVC_WORKER_DB_DSN))
	if err != nil {
		logrus.WithError(err).Fatal("failed to open database")
	}

	store := store.New(db)
	router := http.New(store)
	router.ListenAndServe(os.Getenv(LOCATION_SVC_HTTP_HOST))
}
