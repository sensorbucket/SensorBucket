package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

var (
	// errors
	ErrNoDeviceMatch = errors.New("no device in device service matches EUI of uplink")
)

func NewWorker(processsor processor) *worker {
	// First ensure all the required env variables are present
	w := worker{
		appName:    env.Must("APP_NAME"),
		appType:    env.Must("APP_TYPE"),
		mqQueue:    env.Must("AMQP_QUEUE"),
		mqErrTopic: env.Must("AMQP_ERR_TOPIC"),
		mqHost:     env.Must("AMQP_HOST"),
		mqXchg:     env.Must("AMQP_XCHG"),
		mqPrefetch: env.Could("AMQP_PREFETCH", "5"),
	}

	prefetch, err := strconv.Atoi(w.mqPrefetch)
	if err != nil {
		panic(err)
	}
	conn := mq.NewConnection(w.mqHost)
	publisher := conn.Publisher(w.mqXchg, func(c *amqp091.Channel) error {
		return c.ExchangeDeclare(w.mqXchg, "topic", true, false, false, false, nil)
	})
	consumer := conn.Consume(w.mqQueue, func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(w.mqQueue, true, false, false, false, amqp091.Table{})
		c.Qos(prefetch, 0, true)
		return err
	})
	cancelToken := make(chan any, 1)

	go conn.Start()
	go func(conn *mq.AMQPConnection) {

		// Whenever a value is put in the cancelToken, shutdown the AMQP connnection
		<-cancelToken
		conn.Shutdown()
	}(conn)

	w.processor = processsor
	w.cancelToken = cancelToken
	w.publisher = publisher
	w.consumer = consumer

	return &w
}

// Will run the given processor. Any returned message will be sent to it's next defined step in the pipeline
func (w *worker) Run() {
	// Await any messages that appear on the message queue
	for delivery := range w.consumer {
		var incoming pipeline.Message
		if err := json.Unmarshal(delivery.Body, &incoming); err != nil {
			log.Printf("Error converting delivery: %v\n", err)
			w.publishError(err)
			delivery.Nack(false, false)
			continue
		}

		// Once a message has been received, process it using the worker-unique processor
		result, err := w.processor(incoming)
		if err != nil {
			log.Printf("Error processing delivery: %v\n", err)
			w.publishError(err)
			delivery.Nack(false, false)
			continue
		}

		// If the worker succesfully processed the result, publish it to the next message queue
		topic, err := incoming.NextStep()
		if err != nil {
			delivery.Nack(false, false)
			w.publishError(err)
			continue
		}
		msgJSON, err := json.Marshal(result)
		if err != nil {
			delivery.Nack(false, false)
			w.publishError(fmt.Errorf("could not marshal pipelines message: %w", err))
			continue
		}
		w.publisher <- mq.PublishMessage{Topic: topic, Publishing: amqp091.Publishing{
			Body: msgJSON,
		}}

		// The message was succesfully handled, ack the message.
		delivery.Ack(false)
	}

	// Shutdown the MQ connection
	w.cancelToken <- true
}

func (w *worker) publishError(err error) error {
	errJSON, err := json.Marshal(workerError{
		Origin:     w.appName,
		OriginType: w.appType,
		Error:      err.Error(),
	})
	if err != nil {
		return fmt.Errorf("could not marshal json: %w", err)
	}
	w.publisher <- mq.PublishMessage{Topic: w.mqErrTopic, Publishing: amqp091.Publishing{
		Body: errJSON,
	}}
	return nil
}

type workerError struct {
	Origin     string `json:"origin"`
	OriginType string `json:"originType"`
	Error      string `json:"error"`
}

type worker struct {
	// Worker info
	appName string
	appType string

	// MQ settings
	mqHost     string
	mqQueue    string
	mqErrTopic string
	mqXchg     string
	mqPrefetch string

	processor   processor
	cancelToken chan any
	publisher   publisher
	consumer    consumer
}

type processor func(pipeline.Message) (pipeline.Message, error)
type publisher chan<- mq.PublishMessage
type consumer <-chan amqp091.Delivery
