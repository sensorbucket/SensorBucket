package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/internal/measurements"
)

var (
	ErrTransportClosing = errors.New("transport is shutting down")
	ErrChannelClosed    = errors.New("channel is closed")

	RETRY_DELAY = 5 * time.Second
)

// TransportAMQP provides a transport to interact with the measurement service through AMQP
type TransportAMQP struct {
	svc  *measurements.Service
	xchg string
	q    string

	connection *amqp091.Connection
	channel    *amqp091.Channel

	done chan struct{}
}

// OptsAMQP
type OptsAMQP struct {
	Service  *measurements.Service
	Exchange string
	Queue    string
}

func NewAMQP(opts OptsAMQP) *TransportAMQP {
	return &TransportAMQP{
		svc:  opts.Service,
		xchg: opts.Exchange,
		q:    opts.Queue,
		done: make(chan struct{}),
	}
}

func (t *TransportAMQP) Start(addr string) error {
	log.Printf("Started AMQP transport at %s\n", addr)
	defer log.Println("Stopped AMQP transport")

	for {
		// Connect to AMQP server
		if err := t.connect(addr); err != nil {
			log.Printf("Error connecting to AMQP server: %v\n", err)
			select {
			case <-t.done:
				return nil
			case <-time.After(RETRY_DELAY):
				continue
			}
		}
		defer t.connection.Close()
		log.Println("Connected to AMQP")

		// Create consumer
		consumer, err := startConsumer(t.connection, t.q, t.xchg)
		if err != nil {
			log.Printf("Error creating consumer: %v\n", err)
			continue
		}
		defer consumer.Close()

	process_loop:
		for {
			select {
			case <-t.done:
				return ErrTransportClosing
			case d, ok := <-consumer.Deliveries():
				if !ok {
					// Break the process loops causing a reconnect
					break process_loop
				}
				if err := t.processDelivery(d); err != nil {
					log.Printf("error processing message: %v", err)
					d.Nack(false, false)
					continue
				}
				d.Ack(false)
			}
		}
	}
}

func (t *TransportAMQP) connect(addr string) error {
	if t.channel != nil {
		t.channel.Close()
	}
	if t.connection != nil {
		t.connection.Close()
	}

	// Connect
	conn, err := amqp091.Dial(addr)
	if err != nil {
		return fmt.Errorf("error connecting to AMQP server: %w", err)
	}
	t.connection = conn

	return nil
}

func (t *TransportAMQP) processDelivery(d amqp091.Delivery) error {
	var measurement measurements.IntermediateMeasurement
	if err := json.Unmarshal(d.Body, &measurement); err != nil {
		return fmt.Errorf("error unmarshalling measurement from amqp message: %v", err)
	}

	if err := measurement.Validate(); err != nil {
		return fmt.Errorf("error validating measurement: %v", err)
	}

	if err := t.svc.StoreMeasurement(measurement); err != nil {
		return fmt.Errorf("error storing measurement: %v", err)
	}

	return nil
}

func (t *TransportAMQP) Shutdown() {
	close(t.done)
}
