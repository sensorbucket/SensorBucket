package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/tracing/migrations"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
	tracinginfra "sensorbucket.nl/sensorbucket/services/tracing/tracing/infra"
	tracingtransport "sensorbucket.nl/sensorbucket/services/tracing/tracing/transport"
)

var (
	DB_DSN                      = "host=localhost dbname=tracing user=sensorbucket password=sensorbucket port=5432"
	AMQP_HOST                   = "amqp://guest:guest@localhost:5672/"
	AMQP_QUEUE_PIPELINEMESSAGES = "tracing"
	AMQP_QUEUE_ERRORS           = "errors"
	// DB_DSN                      = env.Must("DB_DSN")
	// AMQP_HOST                   = env.Must("AMQP_HOST")
	// AMQP_QUEUE_PIPELINEMESSAGES = env.Must("AMQP_QUEUE_PIPELINEMESSAGES")
	// AMQP_QUEUE_ERRORS           = env.Must("AMQP_QUEUE_ERRORS")
)

func main() {
	db, err := createDB()
	if err != nil {
		panic(fmt.Sprintf("could not create database connection: %v\n", err))
	}

	// Setup the tracing service
	tracingStepStore := tracinginfra.NewPSQL(db)
	tracingService := tracing.New(tracingStepStore)

	// Setup MQ Transports
	amqpConn := mq.NewConnection(AMQP_HOST)
	tracingtransport.StartMQ(tracingService, amqpConn, AMQP_QUEUE_ERRORS, AMQP_QUEUE_PIPELINEMESSAGES)
	go amqpConn.Start()

	select {}
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
