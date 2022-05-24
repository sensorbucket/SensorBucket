package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"sensorbucket.nl/internal/locations/store"
	"sensorbucket.nl/internal/locations/transport"
)

var (
	LOCATION_SVC_HTTP_HOST     = os.Getenv("LOCATION_SVC_HTTP_HOST")     // :8080
	LOCATION_SVC_WORKER_DB_DSN = os.Getenv("LOCATION_SVC_WORKER_DB_DSN") // postgresql://root:root@localhost:5432/todos?sslmode=disable
)

func main() {
	if err := Run(); err != nil {
		logrus.WithError(err).Fatalln("fatal error occured")
	}
}

func Run() error {
	db, err := sqlx.Open("pgx", LOCATION_SVC_WORKER_DB_DSN)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	store := store.New(db)
	transport := transport.New(store)

	srv := &http.Server{
		Addr:    LOCATION_SVC_HTTP_HOST,
		Handler: cors.AllowAll().Handler(transport),
	}

	errC := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
	}()
	logrus.Infof("HTTP Server listening on %s", srv.Addr)

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-errC:
	case <-sigC:
	}
	logrus.Infoln("Shutting down...")

	srv.Shutdown(context.Background())

	return err
}
