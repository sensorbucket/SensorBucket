package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/tracing/ingress-archiver/service"
	"sensorbucket.nl/sensorbucket/services/tracing/migrations"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
	tracinginfra "sensorbucket.nl/sensorbucket/services/tracing/tracing/infra"
	tracingtransport "sensorbucket.nl/sensorbucket/services/tracing/tracing/transport"
)

var (
	DB_DSN                      = env.Must("DB_DSN")
	AMQP_HOST                   = env.Must("AMQP_HOST")
	AMQP_QUEUE_PIPELINEMESSAGES = env.Must("AMQP_QUEUE_PIPELINEMESSAGES")
	AMQP_QUEUE_ERRORS           = env.Must("AMQP_QUEUE_ERRORS")
	AMQP_QUEUE_INGRESS          = env.Could("AMQP_QUEUE_INGRESS", "archive-ingress")
	AMQP_XCHG_INGRESS           = env.Could("AMQP_XCHG_INGRESS", "ingress")
	AMQP_XCHG_INGRESS_TOPIC     = env.Could("AMQP_XCHG_INGRESS_TOPIC", "ingress.*")
)

func main() {
	// Create shutdown context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := createDB()
	if err != nil {
		panic(fmt.Sprintf("could not create database connection: %v\n", err))
	}

	archiverConn := mq.NewConnection(AMQP_HOST)
	tracingConn := mq.NewConnection(AMQP_HOST)

	go archiverConn.Start()
	go tracingConn.Start()

	// Setup the ingress-archiver service
	go func() {
		store := ingressarchiver.NewStorePSQL(db)
		svc := ingressarchiver.New(store)
		ingressarchiver.StartIngressDTOConsumer(
			archiverConn, svc,
			AMQP_QUEUE_INGRESS, AMQP_XCHG_INGRESS, AMQP_XCHG_INGRESS_TOPIC,
		)
	}()

	// Setup the tracing service
	go func() {
		tracingStepStore := tracinginfra.NewStorePSQL(db)
		tracingService := tracing.New(tracingStepStore)
		tracingtransport.StartMQ(tracingService, tracingConn, AMQP_QUEUE_ERRORS, AMQP_QUEUE_PIPELINEMESSAGES)
	}()

	log.Println("Server running, send interrupt (i.e. CTRL+C) to initiate shutdown")
	<-ctx.Done()
	log.Println("Shutting down... send another interrupt to force shutdown")

	// Create timeout for graceful shutdown
	_, cancelTO := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelTO()

	// Shutdown transports
	archiverConn.Shutdown()
	tracingConn.Shutdown()

	log.Println("Shutdown complete")
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
