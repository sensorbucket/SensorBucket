package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infra"
	devicetransport "sensorbucket.nl/sensorbucket/services/core/devices/transport"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	measurementsinfra "sensorbucket.nl/sensorbucket/services/core/measurements/infra"
	measurementtransport "sensorbucket.nl/sensorbucket/services/core/measurements/transport"
	"sensorbucket.nl/sensorbucket/services/core/migrations"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	processinginfra "sensorbucket.nl/sensorbucket/services/core/processing/infra"
	processingtransport "sensorbucket.nl/sensorbucket/services/core/processing/transport"
)

var (
	DB_DSN                  = env.Must("DB_DSN")
	AMQP_HOST               = env.Must("AMQP_HOST")
	AMQP_QUEUE_MEASUREMENTS = env.Could("AMQP_QUEUE_MEASUREMENTS", "measurements")
	HTTP_ADDR               = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE               = env.Could("HTTP_BASE", "http://localhost:3000/api")
	SYS_ARCHIVE_TIME        = env.Could("SYS_ARCHIVE_TIME", "30")
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
	}
}

func Run() error {
	// Create shutdown context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := createDB()
	if err != nil {
		return fmt.Errorf("could not create database connection: %w", err)
	}

	devicestore := deviceinfra.NewPSQLStore(db)
	sensorGroupStore := deviceinfra.NewPSQLSensorGroupStore(db)
	deviceservice := devices.New(devicestore, sensorGroupStore)
	deviceshttp := devicetransport.NewHTTPTransport(deviceservice, HTTP_BASE)

	sysArchiveTime, err := strconv.Atoi(SYS_ARCHIVE_TIME)
	if err != nil {
		return fmt.Errorf("could not convert SYS_ARCHIVE_TIME to integer: %w", err)
	}
	measurementstore := measurementsinfra.NewPSQL(db)
	measurementservice := measurements.New(measurementstore, sysArchiveTime)
	measurementhttp := measurementtransport.NewHTTP(measurementservice, HTTP_BASE)

	processingstore := processinginfra.NewPSQLStore(db)
	processingservice := processing.New(processingstore)
	processinghttp := processingtransport.NewTransport(processingservice, HTTP_BASE)

	// Setup HTTP Transport
	r := chi.NewRouter()
	deviceshttp.SetupRoutes(r)
	measurementhttp.SetupRoutes(r)
	processinghttp.SetupRoutes(r)
	httpsrv := createHTTPServer(r)
	go httpsrv.ListenAndServe()
	log.Printf("HTTP Listening: %s\n", httpsrv.Addr)

	// Setup MQ Transports
	amqpConn := mq.NewConnection(AMQP_HOST)
	measurementtransport.StartMQ(measurementservice, amqpConn, AMQP_QUEUE_MEASUREMENTS)
	go amqpConn.Start()

	// Wait for shutdown signal
	log.Println("Server running, send interrupt (i.e. CTRL+C) to initiate shutdown")
	<-ctx.Done()
	log.Println("Shutting down... send another interrupt to force shutdown")

	// Create timeout for graceful shutdown
	ctxTO, cancelTO := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelTO()

	// Shutdown transports
	httpsrv.Shutdown(ctxTO)
	amqpConn.Shutdown()

	log.Println("Shutdown complete")
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
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}
	return db, nil
}
