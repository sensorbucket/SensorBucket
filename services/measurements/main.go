package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/cors"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/measurements/migrations"
	"sensorbucket.nl/sensorbucket/services/measurements/service"
	"sensorbucket.nl/sensorbucket/services/measurements/store"
	"sensorbucket.nl/sensorbucket/services/measurements/transport"
)

var (
	HTTP_BASE     = env.Must("HTTP_BASE")
	HTTP_ADDR     = env.Must("HTTP_ADDR")
	DB_DSN        = env.Must("DB_DSN")
	AMQP_HOST     = env.Must("AMQP_HOST")
	AMQP_QUEUE    = env.Must("AMQP_QUEUE")
	AMQP_PREFETCH = env.Could("AMQP_PREFETCH", "5")
)

func main() {
	if err := Run(); err != nil {
		log.Fatalf("Fatal error: %s\n", err)
	}
}

func Run() error {
	mqPrefetch, err := strconv.Atoi(AMQP_PREFETCH)
	if err != nil {
		return fmt.Errorf("MQ_PREFETCH environment variable is not an integer: %v", err)
	}
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
		c.Qos(mqPrefetch, 0, true)
		_, err := c.QueueDeclare(AMQP_QUEUE, true, false, false, false, amqp091.Table{})
		return err
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
