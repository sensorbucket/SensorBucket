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

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
	"sensorbucket.nl/internal/measurements"
	"sensorbucket.nl/internal/measurements/store"
	"sensorbucket.nl/internal/measurements/transport"
)

var (
	MS_HTTP_BASE     = mustEnv("MS_HTTP_BASE")
	MS_HTTP_ADDR     = mustEnv("MS_HTTP_ADDR")
	MS_DB_DSN        = mustEnv("MS_DB_DSN")
	MS_AMQP_URL      = mustEnv("MS_AMQP_URL")
	MS_AMQP_EXCHANGE = mustEnv("MS_AMQP_EXCHANGE")
	MS_AMQP_QUEUE    = mustEnv("MS_AMQP_QUEUE")
)

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("%s environment variable not set", key))
	}
	return val
}

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}
}

func Run() error {
	db, err := sqlx.Open("pgx", MS_DB_DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %s", err)
	}
	store := store.NewPSQL(db)

	svc := measurements.New(store)
	amqpTransport := transport.NewAMQP(transport.OptsAMQP{
		Service:  svc,
		Exchange: MS_AMQP_EXCHANGE,
		Queue:    MS_AMQP_QUEUE,
	})
	defer amqpTransport.Shutdown()

	// Start receiving messages in coroutine
	errC := make(chan error)
	go func() {
		if err := amqpTransport.Start(MS_AMQP_URL); err != nil {
			errC <- fmt.Errorf("failed to connect to AMQP: %s", err)
		}
	}()
	log.Println("Processing started...")
	defer log.Println("Processing stopped...")

	// Start http server
	httpTransport := transport.NewHTTP(svc, MS_HTTP_BASE)
	httpSRV := &http.Server{
		Addr:    MS_HTTP_ADDR,
		Handler: cors.AllowAll().Handler(httpTransport),
	}
	go func() {
		if err := httpSRV.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- fmt.Errorf("http server failed: %s", err)
		}
	}()
	log.Printf("HTTP API started on %s", MS_HTTP_ADDR)
	defer log.Println("HTTP API stopped...")

	// Wait for error or SIGINT
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err = <-errC:
	case <-sigC:
	}

	// Shutdown transports
	amqpTransport.Shutdown()
	httpSRV.Shutdown(context.Background())

	return err
}
