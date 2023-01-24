package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/httpimporter/service"
)

var (
	HTTP_ADDR    = env.Could("HTTP_ADDR", ":3000")
	AMQP_HOST    = env.Must("AMQP_HOST")
	AMQP_XCHG    = env.Must("AMQP_XCHG")
	SVC_PIPELINE = env.Must("SVC_PIPELINE")

	ErrInvalidUUID = web.NewError(
		http.StatusBadRequest,
		"Invalid pipeline UUID provided",
		"ERR_PIPELINE_UUID_INVALID",
	)
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
	}
}

func Run() error {
	// Create AMQP Message Queue
	mq := service.NewAMQPQueue(AMQP_HOST, AMQP_XCHG)
	go mq.Start()
	defer mq.Shutdown()
	log.Printf("AMQP Publisher started...\n")

	// Create pipeline service
	ps := service.NewPipelineServiceHTTP(SVC_PIPELINE)

	// Create http importer service
	svc := service.New(mq, ps)

	// Setup HTTP
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		Handler:      svc,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	log.Printf("HTTP Server listening on: %s\n", srv.Addr)
	return srv.ListenAndServe()
}
