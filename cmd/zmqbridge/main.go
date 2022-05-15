package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sensorbucket.nl/internal/zmqbridge"
)

var (
	AMQP_HOST      = mustEnv("WORKER_AMQP_HOST")
	AMQP_XCHG      = mustEnv("WORKER_AMQP_XCHG")
	AMQP_ROUTE_KEY = mustEnv("WORKER_AMQP_ROUTE_KEY")
	ZMQ_SRC        = mustEnv("WORKER_ZMQ_BIND")
)

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment variable not set", key)
	}
	return val
}

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}

func Run() error {
	zmq := zmqbridge.NewZMQ()
	go zmq.StartConsuming(ZMQ_SRC)
	defer zmq.Shutdown()

	amqp := zmqbridge.NewAMQP(AMQP_XCHG)
	go amqp.Start(AMQP_HOST)
	defer amqp.Shutdown()

	// Catch interrupts
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	//
	amqpC := amqp.Channel()

	// Loop until signal
	for {
		select {
		case msg := <-zmq.Channel():
			amqpC <- msg
		case <-sigC:
			return nil
		}
	}
}
