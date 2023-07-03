package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	tracing "sensorbucket.nl/sensorbucket/services/tracing/service"
)

var (
	REDIS_ADDR = env.Must("REDIS_ADDR")
	AMQP_HOST  = env.Must("AMQP_HOST")
	AMQP_QUEUE = env.Must("AMQP_QUEUE")
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
	}
}

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	rc := redis.NewClient(&redis.Options{
		Addr: REDIS_ADDR,
	})
	// Build Archiver and StateStorer
	archiveTTL := 24 * time.Hour
	stateTTL := 1 * time.Hour

	store := tracing.NewRedisStore(rc, archiveTTL, stateTTL)
	svc := tracing.New(store, store)

	metrics := tracing.NewMetricsService(store)
	prometheus.Register(metrics)

	srv := &http.Server{
		Addr:    ":2112",
		Handler: promhttp.Handler(),
	}
	go func() {
		srv.ListenAndServe()
	}()

	amqp := mq.NewAMQPConsumer(AMQP_HOST, AMQP_QUEUE, func(c *amqp091.Channel) error {
		// TODO: Setup queue, exchange etc
		_, err := c.QueueDeclare(AMQP_QUEUE, false, false, false, false, nil)
		return err
	})
	go amqp.Start()

	go func() {
		for del := range amqp.Consume() {
			svc.ProcessDelivery(del)
		}
	}()
	fmt.Print("Running\n")

	// Wait for interrupt
	<-ctx.Done()
	fmt.Print("Shutting down gracefully\n")
	ctxTO, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	amqp.Shutdown()
	srv.Shutdown(ctxTO)
	return nil
}
