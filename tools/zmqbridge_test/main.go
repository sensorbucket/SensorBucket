package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	zmq "github.com/pebbe/zmq4"
)

var (
	ZMQ_DST      = mustEnv("WORKER_ZMQ_CONNECT")
	flagInterval = flag.Duration("i", time.Second, "interval between messages")
	DEBUG        = false
)

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment variable not set", key)
	}
	return val
}

func main() {
	flag.Parse()
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}

func Run() error {
	ctx, err := zmq.NewContext()
	if err != nil {
		return fmt.Errorf("error creating ZMQ context: %s", err)
	}
	defer ctx.Term()

	sock, err := ctx.NewSocket(zmq.DEALER)
	if err != nil {
		return fmt.Errorf("error creating ZMQ dealer: %s", err)
	}
	if err := sock.Connect(ZMQ_DST); err != nil {
		return fmt.Errorf("error binding ZMQ dealer: %s", err)
	}

	log.Printf("Bridge tester active\n")
	defer log.Printf("Bridge tester stopped\n")

	var id byte = 0
	for {
		id++
		if _, err := sock.SendMessage([][]byte{{id}}, 0); err != nil {
			return fmt.Errorf("error sending ZMQ message: %s", err)
		}
		log.Printf("Sent frame\n")
		<-time.After(*flagInterval)
	}
}
