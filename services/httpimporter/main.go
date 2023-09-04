package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/httpimporter/service"
)

var (
	HTTP_ADDR       = env.Could("HTTP_ADDR", ":3000")
	AMQP_HOST       = env.Could("AMQP_HOST", "amqp://guest:guest@localhost/")
	AMQP_XCHG       = env.Could("AMQP_XCHG", "ingress")
	AMQP_XCHG_TOPIC = env.Could("AMQP_XCHG_TOPIC", "ingress.httpimporter")

	ErrInvalidUUID = web.NewError(
		http.StatusBadRequest,
		"Invalid pipeline UUID provided",
		"ERR_PIPELINE_UUID_INVALID",
	)
)

func main() {
	if err := Run(); err != nil {
		panic(fmt.Sprintf("Fatal error: %v", err))
	}
}

func Run() error {
	// Create AMQP Message Queue
	mqConn := mq.NewConnection(AMQP_HOST)
	go mqConn.Start()
	defer mqConn.Shutdown()
	publisher := service.StartIngressDTOPublisher(mqConn, AMQP_XCHG, AMQP_XCHG_TOPIC)
	log.Printf("AMQP Publisher started...\n")

	// Create http importer service
	svc := service.New(publisher)

	// Setup HTTP
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		Handler:      cors.AllowAll().Handler(svc),
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	log.Printf("HTTP Server listening on: %s\n", srv.Addr)
	return srv.ListenAndServe()
}
