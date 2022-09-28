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
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/cors"
	"sensorbucket.nl/internal/env"
	"sensorbucket.nl/pkg/mq"
	"sensorbucket.nl/services/measurements/migrations"
	"sensorbucket.nl/services/measurements/service"
	"sensorbucket.nl/services/measurements/store"
	"sensorbucket.nl/services/measurements/transport"
)

var (
	HTTP_BASE  = env.Must("HTTP_BASE")
	HTTP_ADDR  = env.Must("HTTP_ADDR")
	DB_DSN     = env.Must("DB_DSN")
	AMQP_HOST  = env.Must("AMQP_URL")
	AMQP_QUEUE = env.Must("AMQP_QUEUE")
)

func main() {
	if err := Run(); err != nil {
		log.Fatalf("Fatal error: %s\n", err)
	}
}

func Run() error {
	db, err := sqlx.Open("pgx", DB_DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	store := store.NewPSQL(db)
	svc := service.New(store)

	// Start receiving messages in coroutine
	consumer := mq.NewAMQPConsumer(AMQP_HOST, AMQP_QUEUE, func(c *amqp091.Channel) error {
		return nil
	})
	go consumer.Start()

	errC := make(chan error)
	mqTransport := transport.NewMQ(svc, consumer)
	go mqTransport.Start()
	log.Println("Processing started...")
	defer log.Println("Processing stopped...")

	// Start http server
	httpTransport := transport.NewHTTP(svc, HTTP_BASE)
	httpSRV := &http.Server{
		Addr:    HTTP_ADDR,
		Handler: cors.AllowAll().Handler(httpTransport),
	}
	go func() {
		if err := httpSRV.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- fmt.Errorf("http server failed: %s", err)
		}
	}()
	log.Printf("HTTP API started on %s\n", HTTP_ADDR)
	defer log.Println("HTTP API stopped...")

	// Wait for error or SIGINT
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err = <-errC:
	case <-sigC:
	}

	// Shutdown transports
	consumer.Shutdown()
	httpSRV.Shutdown(context.Background())

	return err
}
