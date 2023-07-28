package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type Processor func(pipeline.Message) (pipeline.Message, error)

func Run(process Processor) error {
	AMQP_QUEUE := env.Must("AMQP_QUEUE")
	AMQP_HOST := env.Must("AMQP_HOST")
	AMQP_XCHG := env.Must("AMQP_XCHG")
	AMQP_PREFETCH := env.Could("AMQP_PREFETCH", "5")

	prefetch, err := strconv.Atoi(AMQP_PREFETCH)
	if err != nil {
		return err
	}
	mqConn := mq.NewConnection(AMQP_HOST)
	publisher := mqConn.Publisher(AMQP_XCHG, func(c *amqp091.Channel) error {
		return c.ExchangeDeclare(AMQP_XCHG, "topic", true, false, false, false, nil)
	})
	consumer := mqConn.Consume(AMQP_QUEUE, func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(AMQP_QUEUE, true, false, false, false, amqp091.Table{})
		c.Qos(prefetch, 0, true)
		return err
	})
	go mqConn.Start()

	// Process messages
	go startConsuming(consumer, process, publisher)

	// wait for a signal to shutdown
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	<-sigC
	mqConn.Shutdown()
	log.Println("shutting down")
	return nil
}

var ErrNoDeviceMatch = errors.New("no device in device service matches EUI of uplink")

func startConsuming(c <-chan amqp091.Delivery, process Processor, p chan<- mq.PublishMessage) {
	consume := func(delivery amqp091.Delivery) error {
		var err error
		var msg pipeline.Message
		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			return fmt.Errorf("could not unmarshal delivery: %v", err)
		}

		// Do process
		msg, err = process(msg)
		if err != nil {
			return fmt.Errorf("could not process message: %v", err)
		}

		// Publish result
		topic, err := msg.NextStep()
		if err != nil {
			return fmt.Errorf("message has no steps remaining: %w", err)
		}
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("could not marshal pipelines message: %w", err)
		}
		p <- mq.PublishMessage{Topic: topic, Publishing: amqp091.Publishing{
			Body: msgJSON,
		}}
		return nil
	}

	for delivery := range c {
		if err := consume(delivery); err != nil {
			log.Printf("Error processing delivery: %v\n", err)
			delivery.Nack(false, false)
			continue
		}
		delivery.Ack(false)
	}
}
