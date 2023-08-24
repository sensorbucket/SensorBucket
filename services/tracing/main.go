package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	HTTP_ADDR                   = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE                   = env.Could("HTTP_BASE", "http://localhost:3000/api")
	AMQP_HOST                   = env.Could("AMQP_HOST", "")
	AMQP_QUEUE_PIPELINEMESSAGES = env.Could("AMQP_QUEUE_PIPELINEMESSAGES", "")
	AMQP_QUEUE_ERRORS           = env.Could("AMQP_QUEUE_ERRORS", "")
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

	mqConn := mq.NewConnection(AMQP_HOST)
	go mqConn.Start()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Setup the ingress-archiver service
	{
		store := ingressarchiver.NewStorePSQL(db)
		svc := ingressarchiver.New(store)
		go ingressarchiver.StartIngressDTOConsumer(
			mqConn, svc,
			AMQP_QUEUE_INGRESS, AMQP_XCHG_INGRESS, AMQP_XCHG_INGRESS_TOPIC,
		)
		ingressarchiver.CreateHTTPTransport(r, svc)
	}

	// Setup the tracing service
	go func() {
		tracingStepStore := tracinginfra.NewStorePSQL(db)
		tracingService := tracing.New(tracingStepStore)
		go tracingtransport.StartMQ(tracingService, mqConn, AMQP_QUEUE_ERRORS, AMQP_QUEUE_PIPELINEMESSAGES)
		tracinghttp := tracingtransport.NewHTTP(tracingService, HTTP_BASE)
		tracinghttp.SetupRoutes(r)
	}()

	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      r,
	}
	go srv.ListenAndServe()

	log.Println("Server running, send interrupt (i.e. CTRL+C) to initiate shutdown")
	<-ctx.Done()
	log.Println("Shutting down... send another interrupt to force shutdown")

	// Create timeout for graceful shutdown
	_, cancelTO := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelTO()

	// Shutdown transports
	mqConn.Shutdown()
	mqConn.Shutdown()

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
