package mq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/internal/env"
)

var defaultPrefetchCount int = env.CouldInt("AMQP_PREFETCH", 100)

func DefaultPrefetch() int {
	return defaultPrefetchCount
}

type SetupOpts struct {
	Queue    string
	Exchange string
	Topic    string
}

type SetupOption func(c *amqp091.Channel) error

func setupChannel(c *amqp091.Channel, opts []SetupOption) error {
	for _, f := range opts {
		if err := f(c); err != nil {
			return err
		}
	}
	return nil
}

func WithDefaults() SetupOption {
	return func(c *amqp091.Channel) error {
		return c.Qos(defaultPrefetchCount, 0, false)
	}
}

func WithTopicBinding(queue, exchange, topic string) SetupOption {
	return func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(queue, true, false, false, false, amqp091.Table{
			"x-queue-type": "quorum",
		})
		if err != nil {
			return fmt.Errorf("error declaring amqp queue: %w", err)
		}
		err = c.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("error declaring amqp exchange: %w", err)
		}
		err = c.QueueBind(queue, topic, exchange, false, nil)
		if err != nil {
			return fmt.Errorf("error binding amqp queue to exchange: %w", err)
		}
		return nil
	}
}

func WithExchange(exchange string) SetupOption {
	return func(c *amqp091.Channel) error {
		err := c.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("error declaring amqp exchange: %w", err)
		}
		return nil
	}
}
