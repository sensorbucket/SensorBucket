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

var (
	// errors
	ErrNoDeviceMatch = errors.New("no device in device service matches EUI of uplink")

	// env variables
	APP_NAME       = env.Must("APP_NAME")
	APP_TYPE       = env.Must("APP_TYPE")
	AMQP_QUEUE     = env.Must("AMQP_QUEUE")
	AMQP_ERR_TOPIC = env.Must("AMQP_ERR_TOPIC")
	AMQP_HOST      = env.Must("AMQP_HOST")
	AMQP_XCHG      = env.Must("AMQP_XCHG")
	AMQP_PREFETCH  = env.Could("AMQP_PREFETCH", "5")
)

type WorkerError struct {
	Origin     string `json:"origin"`
	OriginType string `json:"originType"`
	Error      string `json:"error"`
}

type Processor func(pipeline.Message) (pipeline.Message, error)

// Will run the given processor. Any returned message will be sent to it's next defined step in the pipeline
// if no steps remain, an error is returned.
func Run(process Processor) error {
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

func startConsuming(c <-chan amqp091.Delivery, process Processor, p chan<- mq.PublishMessage) {
	for delivery := range c {
		if err := consume(delivery, process, p); err != nil {
			log.Printf("Error processing delivery: %v\n", err)
			delivery.Nack(false, false)
			if err = publishError(err, p); err != nil {
				log.Printf("could not publish error: %v\n", err)
			}
			continue
		}
		delivery.Ack(false)
	}
}

func consume(delivery amqp091.Delivery, process Processor, p chan<- mq.PublishMessage) error {
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
		if errors.Is(err, pipeline.ErrMessageNoSteps) {
			// TODO: is this really an error?
			return fmt.Errorf("message has no steps remaining: %w", err)
		}
		return err
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

func publishError(err error, p chan<- mq.PublishMessage) error {
	errJSON, err := json.Marshal(WorkerError{
		Origin:     APP_NAME,
		OriginType: APP_TYPE,
		Error:      err.Error(),
	})
	if err != nil {
		return fmt.Errorf("could not marshal json: %w", err)
	}
	p <- mq.PublishMessage{Topic: AMQP_ERR_TOPIC, Publishing: amqp091.Publishing{
		Body: errJSON,
	}}
	return nil
}
