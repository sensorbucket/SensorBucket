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

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/services/device/migrations"
	"sensorbucket.nl/sensorbucket/services/device/service"
	"sensorbucket.nl/sensorbucket/services/device/store"
)

var (
	DB_DSN    = env.Must("DB_DSN")
	HTTP_ADDR = env.Could("HTTP_ADDR", ":3000")
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
	}
}

func Run() error {
	db := sqlx.MustOpen("pgx", DB_DSN)
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}

	svc := service.New(store.NewPSQLStore(db))
	trsp := service.NewHTTPTransport(svc)
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      trsp,
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
