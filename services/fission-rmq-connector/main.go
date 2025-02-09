package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
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
			log.Printf("[Warn] Cleanup error(s) occured: %s\n", err)
		}
	}()
	if err := Run(cleanup); err != nil {
		log.Fatalf("error occured: %s\n", err)
	}
}

var (
	AMQP_HOST     = env.Must("AMQP_HOST")
	AMQP_QUEUE    = env.Must("QUEUE_NAME")
	AMQP_TOPIC    = env.Must("TOPIC")
	AMQP_XCHG     = env.Must("EXCHANGE")
	HTTP_ENDPOINT = env.Must("HTTP_ENDPOINT")
	MAX_RETRIES   = env.Could("MAX_RETRIES", "3")
	METRICS_ADDR  = env.Could("METRICS_ADDR", ":2112")
)

func Run(cleanup cleanupper.Cleanupper) error {
	maxRetries, err := strconv.Atoi(MAX_RETRIES)
	if err != nil {
		return err
	}

	stopProfiler, err := web.RunProfiler()
	if err != nil {
		fmt.Printf("could not setup profiler server: %s\n", err)
	}
	cleanup.Add(stopProfiler)

	log.Printf("Consuming from queue: %s and producing to exchange: %s\n", AMQP_QUEUE, AMQP_XCHG)
	conn := mq.NewConnection(AMQP_HOST)
	cleanup.Add(func(ctx context.Context) error {
		conn.Shutdown()
		return nil
	})
	successChan := conn.Publisher(AMQP_XCHG, func(c *amqp091.Channel) error {
		return nil
	})
	consumeChan := conn.Consume(AMQP_QUEUE,
		mq.WithDefaults(),
		mq.WithTopicBinding(AMQP_QUEUE, AMQP_XCHG, AMQP_TOPIC),
	)
	go conn.Start()

	connector := Connector{
		Name:       fmt.Sprintf("%s-(%s)", os.Getenv("SOURCE_NAME"), os.Getenv("HOSTNAME")),
		Endpoint:   HTTP_ENDPOINT,
		MaxRetries: maxRetries,
		Result:     successChan,
	}

	go func() {
		for delivery := range consumeChan {
			go connector.handleDelivery(delivery)
		}
	}()

	if METRICS_ADDR != "" {
		go func() {
			srv := &http.Server{
				Addr:         METRICS_ADDR,
				WriteTimeout: 5 * time.Second,
				ReadTimeout:  5 * time.Second,
				Handler:      promhttp.Handler(),
			}
			log.Printf("Metrics server starting at: %s\n", METRICS_ADDR)
			err := srv.ListenAndServe()
			log.Printf("Metrics server stopped: %s\n", err.Error())
		}()
	}

	log.Printf("RabbitMQ-Fission Connector is running...\n")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	<-ctx.Done()
	log.Printf("RabbitMQ-Fission interrupted, shutting down gracefully...\n")

	return nil
}

type Connector struct {
	Name       string
	Endpoint   string
	MaxRetries int
	Result     chan<- mq.PublishMessage
}

func (c *Connector) handleDelivery(delivery amqp091.Delivery) {
	opsProcessed.Inc()

	res, err := doHTTPRequest(delivery.Body, c.Endpoint, c.MaxRetries)
	if err != nil {
		// This is a Function Invocation error or a fatal worker error
		// can't be a business logic error (ie device not found) such an error would be considered
		// a succesful invocation
		opsFails.Inc()
		c.handleError(delivery, err)
	} else {
		opsSuccesful.Inc()
		c.handleSuccess(delivery, res)
	}
}

func (c *Connector) handleError(delivery amqp091.Delivery, err error) {
	// The invocation failed, what to do?
	log.Printf("Invocation error: %v. Redelivering?: %v", err.Error(), !delivery.Redelivered)
	if !delivery.Redelivered {
		if err := delivery.Nack(false, true); err != nil {
			log.Printf("Error Nacking amqp delivery: %v\n", err)
		}
	}
}

func (c *Connector) handleSuccess(delivery amqp091.Delivery, res *http.Response) {
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.handleError(delivery, err)
		return
	}

	// Get response metadata
	topic := res.Header.Get("X-AMQP-Topic")
	if topic == "" {
		c.handleError(delivery, fmt.Errorf("X-AMQP-Topic header must be set after invoking: %s, but wasn't. Handling as failure", HTTP_ENDPOINT))
		return
	}

	// Publish to exchange
	c.Result <- mq.PublishMessage{
		Topic: topic,
		Publishing: amqp091.Publishing{
			MessageId: delivery.MessageId,
			Body:      body,
		},
	}
	if err := delivery.Ack(false); err != nil {
		log.Printf("Error Acking amqp delivery: %v\n", err)
	}
}

func doHTTPRequest(body []byte, endpoint string, retries int) (*http.Response, error) {
	var res *http.Response
	for retry := 0; retry < retries; retry++ {
		req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("error creating invocation for function: %s, error: %w", endpoint, err)
		}
		req.Header.Set("X-AMQP-Topic", AMQP_TOPIC)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Invocation for %s failed with: %v\n", endpoint, err)
			continue
		}
		if res == nil {
			continue
		}
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			return res, nil
		}
		body, _ := io.ReadAll(res.Body)
		log.Printf("Try of invocation failed with status: %d and body:\n%s\n", res.StatusCode, string(body))
	}
	var statusCode int
	if res != nil {
		statusCode = res.StatusCode
	}
	return nil, fmt.Errorf("invocation of %s failed with statuscode: %d", endpoint, statusCode)
}
