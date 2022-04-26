package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/internal/measurements"
	"sensorbucket.nl/internal/measurements/store"
	"sensorbucket.nl/internal/measurements/transport"
)

var (
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
	t := transport.NewAMQP(transport.OptsAMQP{
		Service:  svc,
		Exchange: MS_AMQP_EXCHANGE,
		Queue:    MS_AMQP_QUEUE,
	})
	defer t.Shutdown()

	// Start receiving messages in coroutine
	errC := make(chan error)
	go func() {
		if err := t.Start(MS_AMQP_URL); err != nil {
			errC <- fmt.Errorf("failed to connect to AMQP: %s", err)
		}
	}()
	log.Println("Processing started...")
	defer log.Println("Processing stopped...")

	// Wait for error or SIGINT
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errC:
		return err
	case <-sigC:
		return nil
	}
}
