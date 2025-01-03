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
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"

	"sensorbucket.nl/sensorbucket/internal/buildinfo"
	"sensorbucket.nl/sensorbucket/internal/cleanupper"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/healthchecker"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infra"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	measurementsinfra "sensorbucket.nl/sensorbucket/services/core/measurements/infra"
	"sensorbucket.nl/sensorbucket/services/core/migrations"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	processinginfra "sensorbucket.nl/sensorbucket/services/core/processing/infra"
	coretransport "sensorbucket.nl/sensorbucket/services/core/transport"
)

var (
	DB_DSN                       = env.Must("DB_DSN")
	AMQP_HOST                    = env.Must("AMQP_HOST")
	AMQP_XCHG_INGRESS_TOPIC      = env.Could("AMQP_XCHG_INGRESS_TOPIC", "ingress.*")
	AMQP_XCHG_PIPELINE_MESSAGES  = env.Could("AMQP_XCHG_PIPELINE_MESSAGES", "pipeline.messages")
	AMQP_QUEUE_MEASUREMENTS      = env.Could("AMQP_QUEUE_MEASUREMENTS", "measurements")
	AMQP_XCHG_MEASUREMENTS_TOPIC = env.Could("AMQP_XCHG_MEASUREMENTS_TOPIC", "storage")
	AMQP_QUEUE_INGRESS           = env.Could("AMQP_QUEUE_INGRESS", "core-ingress")
	AMQP_XCHG_INGRESS            = env.Could("AMQP_XCHG_INGRESS", "ingress")
	AMQP_QUEUE_ERRORS            = env.Could("AMQP_QUEUE_ERRORS", "errors")
	HTTP_ADDR                    = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE                    = env.Could("HTTP_BASE", "http://localhost:3000/api")
	AUTH_JWKS_URL                = env.Could("AUTH_JWKS_URL", "http://oathkeeper:4456/.well-known/jwks.json")
	SYS_ARCHIVE_TIME             = env.Could("SYS_ARCHIVE_TIME", "30")
	MEASUREMENT_BATCH_SIZE       = env.CouldInt("MEASUREMENT_BATCH_SIZE", 1024)
	MEASUREMENT_COMMIT_INTERVAL  = env.CouldInt("MEASUREMENT_COMMIT_INTERVAL", 1000)
)

func main() {
	buildinfo.Print()
	cleanup := cleanupper.Create()
	defer func() {
		if err := cleanup.Execute(5 * time.Second); err != nil {
			log.Printf("[Warn] Cleanup error(s) occured: %s\n", err)
		}
	}()
	if err := Run(cleanup); err != nil {
		panic(fmt.Sprintf("Fatal error: %v\n", err))
	}
}

func Run(cleanup cleanupper.Cleanupper) error {
	// Create shutdown context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	stopProfiler, err := web.RunProfiler()
	if err != nil {
		fmt.Printf("could not setup profiler server: %s\n", err)
	}
	cleanup.Add(stopProfiler)

	db, err := createDB()
	if err != nil {
		return fmt.Errorf("could not create database connection: %w", err)
	}

	keyClient := auth.NewJWKSHttpClient(AUTH_JWKS_URL)

	amqpConn := mq.NewConnection(AMQP_HOST)
	cleanup.Add(func(ctx context.Context) error {
		amqpConn.Shutdown()
		return nil
	})

	devicestore := deviceinfra.NewPSQLStore(db)
	sensorGroupStore := deviceinfra.NewPSQLSensorGroupStore(db)
	deviceservice := devices.New(devicestore, sensorGroupStore)

	sysArchiveTime, err := strconv.Atoi(SYS_ARCHIVE_TIME)
	if err != nil {
		return fmt.Errorf("could not convert SYS_ARCHIVE_TIME to integer: %w", err)
	}
	measurementstore := measurementsinfra.NewPSQL(db)
	measurementservice := measurements.New(measurementstore, sysArchiveTime, keyClient)
	cleanup.Add(measurementservice.StartMeasurementBatchStorer(MEASUREMENT_BATCH_SIZE, time.Duration(MEASUREMENT_COMMIT_INTERVAL)*time.Millisecond))

	processingstore := processinginfra.NewPSQLStore(db)
	processingPipelinePublisher := processinginfra.NewPipelineMessagePublisher(amqpConn, AMQP_XCHG_PIPELINE_MESSAGES)
	processingservice := processing.New(processingstore, processingPipelinePublisher, keyClient)

	// Setup MQ Transports
	go mq.StartQueueProcessor(
		amqpConn,
		AMQP_QUEUE_MEASUREMENTS,
		AMQP_XCHG_PIPELINE_MESSAGES,
		AMQP_XCHG_MEASUREMENTS_TOPIC,
		measurements.MQMessageProcessor(measurementservice),
	)
	go mq.StartQueueProcessor(
		amqpConn,
		AMQP_QUEUE_INGRESS,
		AMQP_XCHG_INGRESS,
		AMQP_XCHG_INGRESS_TOPIC,
		processing.MQIngressDTOProcessor(processingservice),
	)
	go amqpConn.Start()

	// Setup HTTP Transport
	httpsrv := createHTTPServer(coretransport.New(
		HTTP_BASE,
		keyClient,
		deviceservice,
		measurementservice,
		processingservice,
	))
	go func() {
		if err := httpsrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) && err != nil {
			fmt.Printf("HTTP Server error: %v\n", err)
		}
	}()
	log.Printf("HTTP Listening: %s\n", httpsrv.Addr)

	healthShutdown := healthchecker.Create().WithEnv().WithMessagQueue(amqpConn).Start(ctx)
	cleanup.Add(healthShutdown)

	// Wait for shutdown signal
	log.Println("Server running, send interrupt (i.e. CTRL+C) to initiate shutdown")
	<-ctx.Done()
	log.Println("Shutting down... send another interrupt to force shutdown")

	// Create timeout for graceful shutdown
	ctxTO, cancelTO := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelTO()

	// Shutdown transports
	if err := httpsrv.Shutdown(ctxTO); err != nil {
		log.Printf("Error shutting down HTTP Server: %v\n", err)
	}
	return nil
}

func createHTTPServer(h http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      cors.AllowAll().Handler(h),
	}
	return srv
}

func createDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", DB_DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(50)
	db.SetConnMaxIdleTime(4 * time.Minute)
	db.SetConnMaxLifetime(0)
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}
	return db, nil
}
