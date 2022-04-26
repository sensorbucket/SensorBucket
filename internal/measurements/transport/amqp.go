package transport

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/internal/measurements"
)

// TransportAMQP provides a transport to interact with the measurement service through AMQP
type TransportAMQP struct {
	svc  *measurements.Service
	xchg string
	q    string

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

func (t *TransportAMQP) Connect(addr string) error {
	// Connect
	conn, err := amqp.Dial(addr)
	if err != nil {
		return fmt.Errorf("error connecting to AMQP server: %w", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("error creating AMQP channel: %w", err)
	}

	// Declare exchange and queue
	if err := ch.ExchangeDeclare(t.xchg, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("error declaring exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(t.q, true, false, false, false, nil); err != nil {
		return fmt.Errorf("error declaring queue: %w", err)
	}
	if err := ch.QueueBind(t.q, "#", t.xchg, false, nil); err != nil {
		return fmt.Errorf("error binding queue: %w", err)
	}

	// Start receiving messages
	c, err := ch.Consume(t.q, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error consuming queue: %w", err)
	}

	for {
		select {
		// Shutdown
		case <-t.done:
			return nil

		// Process message
		case msg := <-c:
			{
				var measurement measurements.Measurement
				if err := json.Unmarshal(msg.Body, &measurement); err != nil {
					log.Printf("Error unmarshalling measurement from amqp message: %v", err)
					msg.Ack(false)
					continue
				}

				if err := measurement.Validate(); err != nil {
					log.Printf("Error validating measurement: %v", err)
					msg.Ack(false)
					continue
				}

				if err := t.svc.StoreMeasurement(&measurement); err != nil {
					log.Printf("Error storing measurement: %v", err)
					msg.Ack(false)
					continue
				}
			}
		}
	}
}

func (t *TransportAMQP) Shutdown() {
	close(t.done)
}
