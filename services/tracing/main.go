package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/redis/go-redis/v9"
	"sensorbucket.nl/sensorbucket/internal/env"
	tracing "sensorbucket.nl/sensorbucket/services/tracing/service"
)

var (
	REDIS_ADDR = env.Must("REDIS_ADDR")
	AMQP_HOST  = env.Must("AMQP_HOST")
	AMQP_QUEUE = env.Must("AMQP_QUEUE")
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	rc := redis.NewClient(&redis.Options{
		Addr: REDIS_ADDR,
	})
	// Build Archiver and StateStorer
	store := tracing.NewRedisStore(rc)

	if err := tracing.Run(ctx, AMQP_HOST, AMQP_QUEUE, store, store); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
	}
}
