package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/services/pipeline/migrations"
	"sensorbucket.nl/sensorbucket/services/pipeline/service"
	"sensorbucket.nl/sensorbucket/services/pipeline/store"
)

var (
	DB_DSN    = env.Must("DB_DSN")
	HTTP_ADDR = env.Could("HTTP_ADDR", ":3000")
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
	}
}

func Run() error {
	db := sqlx.MustOpen("pgx", DB_DSN)
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}
	svc := service.New(store.NewPSQLStore(db))
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      svc,
	}

	errC := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
	}()

	log.Printf("HTTP Server listening on %s\n", srv.Addr)

	// Wait for error or interrupt signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT)
	var err error
	select {
	case err = <-errC:
	case <-sigC:
	}
	log.Println("Shutting down...")

	srv.Shutdown(context.Background())

	return err
}
