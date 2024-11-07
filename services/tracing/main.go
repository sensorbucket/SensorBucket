package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/tracing/ingress-archiver/service"
	"sensorbucket.nl/sensorbucket/services/tracing/migrations"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
	tracinginfra "sensorbucket.nl/sensorbucket/services/tracing/tracing/infra"
	tracingtransport "sensorbucket.nl/sensorbucket/services/tracing/tracing/transport"
)

var (
	HTTP_ADDR                        = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE                        = env.Could("HTTP_BASE", "http://localhost:3000/api")
	DB_DSN                           = env.Must("DB_DSN")
	AMQP_HOST                        = env.Must("AMQP_HOST")
	AMQP_QUEUE_PIPELINEMESSAGES      = env.Could("AMQP_QUEUE_PIPELINEMESSAGES", "tracing_pipeline_messages")
	AMQP_XCHG_PIPELINEMESSAGES       = env.Could("AMQP_XCHG_PIPELINEMESSAGES", "pipeline.messages")
	AMQP_XCHG_PIPELINEMESSAGES_TOPIC = env.Could("AMQP_XCHG_PIPELINEMESSAGES_TOPIC", "#")
	AMQP_QUEUE_INGRESS               = env.Could("AMQP_QUEUE_INGRESS", "archive-ingress")
	AMQP_XCHG_INGRESS                = env.Could("AMQP_XCHG_INGRESS", "ingress")
	AMQP_XCHG_INGRESS_TOPIC          = env.Could("AMQP_XCHG_INGRESS_TOPIC", "ingress.*")
	AUTH_JWKS_URL                    = env.Could("AUTH_JWKS_URL", "http://oathkeeper:4456/.well-known/jwks.json")
)

func main() {
	// Create shutdown context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	stopProfiler, err := web.RunProfiler()
	if err != nil {
		fmt.Printf("could not setup profiler server: %s\n", err)
	}

	db, err := createDB()
	if err != nil {
		panic(fmt.Sprintf("could not create database connection: %v\n", err))
	}

	mqConn := mq.NewConnection(AMQP_HOST)
	go mqConn.Start()

	r := chi.NewRouter()
	r.Use(
		chimw.Logger,
		auth.Authenticate(auth.NewJWKSHttpClient(AUTH_JWKS_URL)),
		auth.Protect(),
	)

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
	{
		tracingStepStore := tracinginfra.NewStorePSQL(db)
		tracingService := tracing.New(tracingStepStore)
		go tracingtransport.StartMQ(
			tracingService,
			mqConn,
			AMQP_QUEUE_PIPELINEMESSAGES,
			AMQP_XCHG_PIPELINEMESSAGES,
			AMQP_XCHG_PIPELINEMESSAGES_TOPIC,
		)
		tracinghttp := tracingtransport.NewHTTP(tracingService, HTTP_BASE)
		tracinghttp.SetupRoutes(r)
	}

	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      r,
	}
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Printf("HTTP Server error: %v\n", err)
		}
	}()

	log.Println("Server running, send interrupt (i.e. CTRL+C) to initiate shutdown")
	<-ctx.Done()
	log.Println("Shutting down... send another interrupt to force shutdown")

	// Create timeout for graceful shutdown
	ctxTO, cancelTO := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelTO()

	// Shutdown transports
	if err := srv.Shutdown(ctxTO); err != nil {
		log.Printf("error shutting down httpserver: %s\n", err)
	}
	mqConn.Shutdown()
	stopProfiler(ctxTO)

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
