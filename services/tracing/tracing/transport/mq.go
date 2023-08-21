package tracingtransport

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func StartMQ(svc *tracing.Service, conn *mq.AMQPConnection, errQueue string, queue string) func() {
	done := make(chan struct{})
	pipelineMessages := mq.Consume(conn, queue, setupFunc(queue))
	pipelineErrors := mq.Consume(conn, errQueue, setupFunc(errQueue))

	log.Println("Measurement MQ Transport running")
	go process(queue, pipelineMessages, svc.HandlePipelineMessage)
	go process(errQueue, pipelineErrors, svc.HandlePipelineError)

	return func() {
		close(done)
	}
}

func process[T interface{}](queue string, deliveries <-chan amqp091.Delivery, handler func(T) error) {
	log.Println("Measurement MQ Transport running, tracing pipeline errors...")
	for msg := range deliveries {
		var res T
		if err := json.Unmarshal(msg.Body, &res); err != nil {
			msg.Nack(false, false)
			log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
			continue
		}

		if err := handler(res); err != nil {
			msg.Nack(false, false)
			log.Printf("Error handling pipeline message: %v\n", err)
			continue
		}
		msg.Ack(false)
	}
}

func setupFunc(queue string) mq.AMQPSetupFunc {
	return func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(queue, true, false, false, false, nil)
		return err
	}
}
