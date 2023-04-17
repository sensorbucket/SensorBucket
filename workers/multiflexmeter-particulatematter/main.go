package main

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

var (
	AMQP_QUEUE    = env.Must("AMQP_QUEUE")
	AMQP_HOST     = env.Must("AMQP_HOST")
	AMQP_XCHG     = env.Must("AMQP_XCHG")
	AMQP_PREFETCH = env.Could("AMQP_PREFETCH", "5")

	ErrSensorNotFound = errors.New("sensor not found")
)

func main() {
	if err := Run(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

func Run() error {
	prefetch, err := strconv.Atoi(AMQP_PREFETCH)
	if err != nil {
		return err
	}
	publisher := mq.NewAMQPPublisher(AMQP_HOST, AMQP_XCHG, func(c *amqp091.Channel) error {
		return c.ExchangeDeclare(AMQP_XCHG, "topic", true, false, false, false, nil)
	})
	go publisher.Start()

	consumer := mq.NewAMQPConsumer(AMQP_HOST, AMQP_QUEUE, func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(AMQP_QUEUE, true, false, false, false, amqp091.Table{})
		c.Qos(prefetch, 0, true)
		return err
	})
	go consumer.Start()

	// Process messages
	ch := consumer.Consume()
	go processDelivery(ch, publisher)

	// wait for a signal to shutdown
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	<-sigC
	consumer.Shutdown()
	publisher.Shutdown()
	log.Println("shutting down")
	return nil
}

func processDelivery(c <-chan amqp091.Delivery, p *mq.AMQPPublisher) {
	process := func(delivery amqp091.Delivery) error {
		var err error
		var msg pipeline.Message
		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			return fmt.Errorf("could not unmarshal delivery: %v", err)
		}

		// Do process
		msg, err = processMessage(msg)
		if err != nil {
			return fmt.Errorf("could not process message: %w", err)
		}

		// Publish result
		topic, err := msg.NextStep()
		if err != nil {
			return fmt.Errorf("could not get next step: %w", err)
		}
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("could not marshal pipelines message: %w", err)
		}
		p.Publish(topic, amqp091.Publishing{
			Body: msgJSON,
		})
		return nil
	}

	for delivery := range c {
		if err := process(delivery); err != nil {
			log.Printf("Error processing delivery: %v\n", err)
			delivery.Nack(false, false)
			continue
		}
		delivery.Ack(false)
	}
}

func processMessage(msg pipeline.Message) (pipeline.Message, error) {
	data := msg.Payload
	if len(data) == 0 {
		return msg, nil
	}

	// Check if data length is a multiple of 2
	if len(data)%2 != 0 {
		return msg, errors.New("incorrect payload length")
	}

	// Process measurements
	for i := 0; i < len(data); i += 2 {
		measurement := int16(data[i])<<8 | int16(data[i+1])
		err := msg.NewMeasurement().SetSensor("0").SetValue(float64(measurement), "pm_2.5", "ug/m3").Add()
		if err != nil {
			return msg, err
		}
	}

	return msg, nil
}
