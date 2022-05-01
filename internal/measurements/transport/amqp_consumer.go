package transport

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// consumer is a simple AMQP consumer that receives messages from the AMQP server
type consumer struct {
	ch         *amqp091.Channel
	deliveries <-chan amqp091.Delivery
}

func startConsumer(conn *amqp091.Connection, queue, exchange string) (*consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare exchange and queue
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("error declaring exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("error declaring queue: %w", err)
	}
	if err := ch.QueueBind(queue, "#", exchange, false, nil); err != nil {
		return nil, fmt.Errorf("error binding queue: %w", err)
	}

	// Start receiving messages
	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error consuming queue: %w", err)
	}

	return &consumer{
		ch:         ch,
		deliveries: deliveries,
	}, nil
}

func (c *consumer) Deliveries() <-chan amqp091.Delivery {
	return c.deliveries
}

func (c *consumer) Close() error {
	return c.ch.Close()
}
