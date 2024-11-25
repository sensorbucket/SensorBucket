package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	mqqtauth "github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/healthchecker"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	"sensorbucket.nl/sensorbucket/services/mqtt-ingress/service"
)

var (
	APIKEY_TRADE_URL = env.Must("APIKEY_TRADE_URL")
	AMQP_HOST        = env.Could("AMQP_HOST", "amqp://guest:guest@localhost/")
	AMQP_XCHG        = env.Could("AMQP_XCHG", "ingress")
	AMQP_XCHG_TOPIC  = env.Could("AMQP_XCHG_TOPIC", "ingress.httpimporter")
	METRICS_ADDR     = env.Could("METRICS_ADDR", ":2112")
	MQTT_ADDR        = env.Could("MQTT_ADDR", ":1883")
)

type ShutdownFunc func(context.Context) error

type Cleanupper []ShutdownFunc

func (c *Cleanupper) Add(fn ShutdownFunc) {
	*c = append(*c, fn)
}

func (c *Cleanupper) Execute(timeout time.Duration) error {
	ctxTO, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cleanupErrors error
	for _, fn := range *c {
		cleanupErrors = errors.Join(cleanupErrors, fn(ctxTO))
	}

	return cleanupErrors
}

func main() {
	var cleanup Cleanupper
	defer func() {
		if err := cleanup.Execute(5 * time.Second); err != nil {
			log.Printf("[Warn] Cleanup error(s) occured: %s\n", err)
		}
	}()
	if err := Run(cleanup); err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}
}

func Run(cleanup Cleanupper) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	errC := make(chan error, 1)

	stopProfiler, err := web.RunProfiler()
	if err != nil {
		fmt.Printf("could not setup profiler server: %s\n", err)
	}
	cleanup.Add(func(ctx context.Context) error {
		stopProfiler(ctx)
		return nil
	})

	health := healthchecker.Create().WithEnv()

	err = startTelemetry(ctx, cleanup)
	if err != nil {
		return err
	}
	publisher, err := startDTOPublisher(health, cleanup)
	if err != nil {
		return err
	}
	err = startMQTTServer(ctx, publisher, errC, cleanup)
	if err != nil {
		return err
	}
	cleanup.Add(health.Start(ctx))

	// Wait for error or interrupt to stop the server
	log.Println("Server running")
	select {
	case err = <-errC:
		log.Println("Closing due to error")
		break
	case <-ctx.Done():
	}

	// Cleanupper is called after Run

	log.Println("Shutting down...")
	return err
}

func startDTOPublisher(health *healthchecker.Builder, cleanup Cleanupper) (chan<- processing.IngressDTO, error) {
	mqConn := mq.NewConnection(AMQP_HOST)
	health.WithMessagQueue(mqConn)
	go mqConn.Start()
	cleanup.Add(func(_ context.Context) error {
		mqConn.Shutdown()
		return nil
	})
	publisher := service.StartIngressDTOPublisher(mqConn, AMQP_XCHG, AMQP_XCHG_TOPIC)
	log.Println("AMQP Publisher started")
	return publisher, nil
}

func startMQTTServer(ctx context.Context, publisher chan<- processing.IngressDTO, errC chan<- error, cleanup Cleanupper) error {
	authRules := &mqqtauth.Ledger{
		ACL: mqqtauth.ACLRules{
			{Filters: mqqtauth.Filters{
				"#": mqqtauth.WriteOnly,
			}},
		},
	}

	server := mqtt.New(nil)
	if err := server.AddHook(new(service.MQTTProcessor), &service.MQTTProcessorOptions{
		Context:   ctx,
		Publisher: publisher,
		APIKeyTrader: func(apiKey string) (string, error) {
			req, _ := http.NewRequest("GET", APIKEY_TRADE_URL, nil)
			req.Header.Set("Authorization", "Bearer "+apiKey)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", err
			}
			if res.StatusCode != http.StatusOK {
				return "", fmt.Errorf("expected status 200 got: %d", res.StatusCode)
			}
			newAuth, _ := auth.StripBearer(res.Header.Get("Authorization"))
			return newAuth, nil
		},
	}); err != nil {
		return err
	}
	err := server.AddHook(new(mqqtauth.Hook), &mqqtauth.Options{
		Ledger: authRules,
	})
	if err != nil {
		return err
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "t1",
		Address: MQTT_ADDR,
	})
	if err := server.AddListener(tcp); err != nil {
		return err
	}
	go func() {
		if err := server.Serve(); err != nil {
			errC <- err
		}
	}()
	log.Println("MQTT Server started")

	cleanup.Add(func(_ context.Context) error {
		return server.Close()
	})

	return nil
}

func startTelemetry(_ context.Context, cleanup Cleanupper) error {
	promMetrics, err := prometheus.New()
	if err != nil {
		return err
	}
	metricProvider := metric.NewMeterProvider(metric.WithReader(promMetrics))
	cleanup.Add(func(ctx context.Context) error {
		return metricProvider.Shutdown(ctx)
	})
	otel.SetMeterProvider(metricProvider)

	srv := &http.Server{
		Addr:         METRICS_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      promhttp.Handler(),
	}
	cleanup.Add(func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})
	go func() {
		log.Printf("Metrics server starting at: %s\n", METRICS_ADDR)
		err := srv.ListenAndServe()
		log.Printf("Metrics server stopped: %s\n", err.Error())
	}()

	return nil
}
