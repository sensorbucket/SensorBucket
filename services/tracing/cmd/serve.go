package cmd

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
	"github.com/rabbitmq/amqp091-go"
	"github.com/urfave/cli/v2"

	"sensorbucket.nl/sensorbucket/internal/buildinfo"
	"sensorbucket.nl/sensorbucket/internal/cleanupper"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/healthchecker"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/tracing/migrations"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

var (
	HTTP_ADDR                        = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE                        = env.Could("HTTP_BASE", "http://localhost:3000/api")
	DB_DSN                           = env.Must("DB_DSN")
	AMQP_HOST                        = env.Must("AMQP_HOST")
	AMQP_QUEUE_TRACES                = env.Could("AMQP_QUEUE_TRACES", "tracing.traces")
	AMQP_XCHG_PIPELINEMESSAGES       = env.Could("AMQP_XCHG_PIPELINEMESSAGES", "pipeline.messages")
	AMQP_XCHG_PIPELINEMESSAGES_TOPIC = env.Could("AMQP_XCHG_PIPELINEMESSAGES_TOPIC", "#")
	AMQP_QUEUE_INGRESS               = env.Could("AMQP_QUEUE_INGRESS", "archive-ingress")
	AMQP_XCHG_INGRESS                = env.Could("AMQP_XCHG_INGRESS", "ingress")
	AMQP_XCHG_INGRESS_TOPIC          = env.Could("AMQP_XCHG_INGRESS_TOPIC", "ingress.*")
	AUTH_JWKS_URL                    = env.Could("AUTH_JWKS_URL", "http://oathkeeper:4456/.well-known/jwks.json")
)

func cmdServe(cmd *cli.Context) error {
	buildinfo.Print()
	cleanup := cleanupper.Create()
	defer func() {
		if err := cleanup.Execute(5 * time.Second); err != nil {
			log.Printf("[Warn] Cleanup error(s) occured: %s\n", err)
		}
	}()

	// Create shutdown context
	ctx, cancel := signal.NotifyContext(cmd.Context, os.Interrupt)
	defer cancel()

	stopProfiler, err := web.RunProfiler()
	if err != nil {
		log.Printf("could not setup profiler server: %s\n", err)
	}
	cleanup.Add(stopProfiler)

	db, err := createDB()
	if err != nil {
		return fmt.Errorf("could not create database connection: %w", err)
	}

	r := chi.NewRouter()
	r.Use(
		chimw.Logger,
		auth.Authenticate(auth.NewJWKSHttpClient(AUTH_JWKS_URL)),
		auth.Protect(),
	)

	mqConn := mq.NewConnection(AMQP_HOST)
	cleanup.Add(func(ctx context.Context) error {
		mqConn.Shutdown()
		return nil
	})
	go mqConn.Start()

	svc := tracing.Create(db)
	transportHTTP := tracing.CreateTransport(svc)
	r.Mount("/", transportHTTP)
	go mq.StartQueueProcessor(mqConn,
		AMQP_QUEUE_INGRESS, AMQP_XCHG_INGRESS, AMQP_XCHG_INGRESS_TOPIC,
		func() mq.ProcessorFunc {
			return func(delivery amqp091.Delivery) error {
				if err := tracing.ProcessIngress(svc, &delivery); err != nil {
					return fmt.Errorf("process ingress message: %w", err)
				}
				return nil
			}
		},
	)
	go mq.StartQueueProcessor(mqConn,
		AMQP_QUEUE_TRACES, AMQP_XCHG_PIPELINEMESSAGES, AMQP_XCHG_PIPELINEMESSAGES_TOPIC,
		func() mq.ProcessorFunc {
			return func(delivery amqp091.Delivery) error {
				if err := tracing.ProcessMessage(svc, &delivery, "errors"); err != nil {
					return fmt.Errorf("process ingress message: %w", err)
				}
				return nil
			}
		},
	)

	healthShutdown := healthchecker.Create().WithEnv().WithMessagQueue(mqConn).Start(ctx)
	cleanup.Add(healthShutdown)

	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      r,
	}
	cleanup.Add(func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Printf("HTTP Server error: %v\n", err)
		}
	}()

	log.Println("Server running, send interrupt (i.e. CTRL+C) to initiate shutdown")
	<-ctx.Done()
	log.Println("Shutting down... send another interrupt to force shutdown")

	return nil
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
