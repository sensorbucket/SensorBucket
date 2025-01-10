package mq

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/rabbitmq/amqp091-go"
)

var defaultPrefetchCount int = 50

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

func init() {
	prefetchStr, ok := os.LookupEnv("AMQP_PREFETCH")
	if ok {
		prefetch, err := strconv.Atoi(prefetchStr)
		if err != nil {
			log.Fatalf("AMQP_PREFETCH env set but not a number: %s\n", err.Error())
		}
		defaultPrefetchCount = prefetch
	}
}
