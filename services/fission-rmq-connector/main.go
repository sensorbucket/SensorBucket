package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/internal/buildinfo"
	"sensorbucket.nl/sensorbucket/internal/cleanupper"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/mq"
)

var (
	logger = slog.Default()

	opsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sensorbucket_fissionrmqconnect_processed",
		Help: "total number of message queue messages received",
	})
	opsSuccesful = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sensorbucket_fissionrmqconnect_success",
		Help: "total number of message queue messages succesfully processed",
	})
	opsFails = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sensorbucket_fissionrmqconnect_fail",
		Help: "total number of message queue messages failed to process",
	})
)

func main() {
	buildinfo.Print()
	cleanup := cleanupper.Create()
	defer func() {
		if err := cleanup.Execute(5 * time.Second); err != nil {
			logger.Warn("Cleanup error(s) occured", "error", err)
		}
	}()
	if err := Run(cleanup); err != nil {
		logger.Error("error occured", "error", err)
	}
}

var (
	AMQP_HOST     = env.Must("AMQP_HOST")
	AMQP_QUEUE    = env.Must("QUEUE_NAME")
	AMQP_TOPIC    = env.Must("TOPIC")
	AMQP_XCHG     = env.Must("EXCHANGE")
	HTTP_ENDPOINT = env.Must("HTTP_ENDPOINT")
	MAX_RETRIES   = env.CouldInt("MAX_RETRIES", 3)
	METRICS_ADDR  = env.Could("METRICS_ADDR", ":2112")
)

func Run(cleanup cleanupper.Cleanupper) error {
	stopProfiler, err := web.RunProfiler()
	if err != nil {
		logger.Warn("could not setup profiler server", "error", err)
	}
	cleanup.Add(stopProfiler)

	if err := setupMetrics(cleanup); err != nil {
		return err
	}

	logger.Info("Establishing Message Queue connection", "queue", AMQP_QUEUE, "exchange", AMQP_XCHG)
	conn := mq.NewConnection(AMQP_HOST)
	cleanup.Add(func(ctx context.Context) error {
		conn.Shutdown()
		return nil
	})
	go conn.Start()
	publish := conn.Publisher(AMQP_XCHG)
	go mq.StartQueueProcessor(conn, AMQP_QUEUE, AMQP_XCHG, AMQP_TOPIC, buildProcessor(publish))

	logger.Info("Connector is ready")
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()
	logger.Info("Connector shutting down")
	return nil
}

func setupMetrics(cleanup cleanupper.Cleanupper) error {
	if METRICS_ADDR == "" {
		return nil
	}
	srv := &http.Server{
		Addr:         METRICS_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      promhttp.Handler(),
	}
	logger.Info("Metrics server starting", "address", METRICS_ADDR)
	go func() {
		err := srv.ListenAndServe()
		logger.Info("Metrics server stopped", "error", err)
	}()
	cleanup.Add(func(ctx context.Context) error {
		if err := srv.Shutdown(ctx); err != nil {
			logger.Warn("error shutting down metrics server", "error", err)
		}
		return nil
	})
	return nil
}

func buildProcessor(publish chan<- mq.PublishMessage) mq.ProcessorFuncBuilder {
	return func() mq.ProcessorFunc {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		return func(incoming amqp091.Delivery) error {
			opsReceived.Inc()
			err := processIncoming(logger, client, incoming, publish)
			if err != nil {
				opsFails.Inc()
			} else {
				opsSuccesful.Inc()
			}
			return err
		}
	}
}

func processIncoming(logger *slog.Logger, client *http.Client, incoming amqp091.Delivery, publish chan<- mq.PublishMessage) error {
	req, err := http.NewRequest("POST", HTTP_ENDPOINT, bytes.NewReader(incoming.Body))
	if err != nil {
		return fmt.Errorf("could not create request for fission worker: %w", err)
	}
	req.Header.Set("X-AMQP-Topic", AMQP_TOPIC)
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not call fission worker: %w", err)
	}
	if res == nil {
		return fmt.Errorf("could not call fission worker, response is nil?")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Warn("Could not close response body", "error", err)
		}
	}()

	// Do something based on statuscode
	if res.StatusCode < 200 || res.StatusCode > 299 {
		responseMessage := "<none>"
		body, err := io.ReadAll(res.Body)
		if err != nil {
			responseMessage = string(body)
		}
		logger.Warn("Fission worker returned non-ok status", "status_code", res.StatusCode, "body", responseMessage)
		err = fmt.Errorf("response was non-ok status: %d", res.StatusCode)
		// Mark malformed if a bad request was made, this way it isnt requeued
		if res.StatusCode > 399 && res.StatusCode < 500 {
			err = errors.Join(mq.ErrMalformed, err)
		}
		return err
	}

	//
	// Success
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("could not read response body: %w", err)
	}

	// Get response metadata
	topic := res.Header.Get("X-AMQP-Topic")
	if topic == "" {
		return fmt.Errorf("response is missing X-AMQP-Topic header")
	}

	// Publish to exchange
	publish <- mq.PublishMessage{
		Topic: topic,
		Publishing: amqp091.Publishing{
			MessageId: incoming.MessageId,
			Body:      body,
		},
	}
	return nil
}
