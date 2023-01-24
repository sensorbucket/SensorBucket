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
			return fmt.Errorf("could not process message: %v", err)
		}

		// Publish result
		topic, err := msg.NextStep()
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("could not marshal pipelines message: %v", err)
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
	if len(data) < 2 || data[0] != 0x6c || data[1] != 0x11 {
		return msg, errors.New("incorrect payload header")
	}

	// Get first measurement
	millivolt, columnMeters, err := valueToMeasurements(data[2:])
	if err != nil {
		return msg, err
	}
	err = msg.NewMeasurement().SetSensor("0").SetValue(millivolt, "millivolt").Add()
	if err != nil {
		return msg, err
	}
	err = msg.NewMeasurement().SetSensor("0").SetValue(columnMeters, "watercolumn_meters").Add()
	if err != nil {
		return msg, err
	}

	// First bit indicates if there is another measurement appended
	if data[2]&0x80 > 0 {
		millivolt, columnMeters, err := valueToMeasurements(data[5:])
		err = msg.NewMeasurement().SetSensor("1").SetValue(millivolt, "millivolt").Add()
		if err != nil {
			return msg, err
		}
		err = msg.NewMeasurement().SetSensor("1").SetValue(columnMeters, "watercolumn_meters").Add()
		if err != nil {
			return msg, err
		}
	}

	return msg, nil
}

func valueToMeasurements(data []byte) (millivolts, meters float64, err error) {
	if len(data) < 3 {
		err = errors.New("incorrect payload size")
		return
	}

	millivolts = float64((uint32(data[0])<<16)|(uint32(data[1])<<8)|uint32(data[2])) / 100
	meters = 0.102564 * (7.02 + millivolts) / 100.0
	return
}
