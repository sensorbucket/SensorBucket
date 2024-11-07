package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/cors"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/health"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/httpimporter/service"
)

var (
	HTTP_ADDR       = env.Could("HTTP_ADDR", ":3000")
	HEALTH_ADDR     = env.Could("HEALTH_ADDR", ":3030")
	AMQP_HOST       = env.Could("AMQP_HOST", "amqp://guest:guest@localhost/")
	AMQP_XCHG       = env.Could("AMQP_XCHG", "ingress")
	AMQP_XCHG_TOPIC = env.Could("AMQP_XCHG_TOPIC", "ingress.httpimporter")
	AUTH_JWKS_URL   = env.Could("AUTH_JWKS_URL", "http://oathkeeper:4456/.well-known/jwks.json")

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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	stopProfiler, err := web.RunProfiler()
	if err != nil {
		fmt.Printf("could not setup profiler server: %s\n", err)
	}

	// Create AMQP Message Queue
	mqConn := mq.NewConnection(AMQP_HOST)
	go mqConn.Start()
	defer mqConn.Shutdown()
	publisher := service.StartIngressDTOPublisher(mqConn, AMQP_XCHG, AMQP_XCHG_TOPIC)
	log.Printf("AMQP Publisher started...\n")

	// Create http importer service
	jwks := auth.NewJWKSHttpClient(AUTH_JWKS_URL)
	svc := service.New(publisher, jwks)

	// Setup HTTP
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		Handler:      cors.AllowAll().Handler(svc),
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	shutdownHealthEndpoint := health.NewHealthEndpoint().
		WithReadyChecks(
			map[string]health.Check{
				"mqconn-ready": mqConn.Ready,
			},
		).
		WithLiveChecks(
			map[string]health.Check{
				"mqconn-healthy": mqConn.Healthy,
			},
		).
		RunAsServer(HEALTH_ADDR)

	errC := make(chan error)
	go func() {
		log.Printf("HTTP Server listening on: %s\n", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) && err != nil {
			errC <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err = <-errC:
	}

	ctxTO, cancelTO := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTO()

	if err := srv.Shutdown(ctxTO); err != nil {
		log.Printf("Error shutting down HTTP Server: %v\n", err)
	}
	if err := shutdownHealthEndpoint(ctxTO); err != nil {
		log.Printf("Error shutting down Health Server: %v\n", err)
	}
	stopProfiler(ctxTO)

	return err
}
