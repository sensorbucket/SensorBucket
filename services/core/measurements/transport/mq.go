package measurementtransport

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func StartMQ(
	svc *measurements.Service,
	conn *mq.AMQPConnection,
	pipelineMessagesExchange,
	measurementQueue,
	measurementStorageTopic,
	measurementErrorTopic string,
	prefetch int,
) func() {
	done := make(chan struct{})
	consume := mq.Consume(conn, measurementQueue, func(c *amqp091.Channel) error {
		if err := c.Qos(prefetch, 0, false); err != nil {
			return fmt.Errorf("error setting Qos with prefetch on amqp: %w", err)
		}
		q, err := c.QueueDeclare(measurementQueue, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("error declaring amqp queue: %w", err)
		}
		err = c.ExchangeDeclare(pipelineMessagesExchange, "topic", true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("error declaring amqp exchange: %w", err)
		}
		err = c.QueueBind(q.Name, measurementStorageTopic, pipelineMessagesExchange, false, nil)
		if err != nil {
			return fmt.Errorf("error binding amqp queue to exchange: %w", err)
		}
		return nil
	})
	publish := mq.Publisher(conn, pipelineMessagesExchange, func(c *amqp091.Channel) error {
		err := c.ExchangeDeclare(pipelineMessagesExchange, "topic", true, false, false, false, nil)
		return err
	})

	go func() {
		log.Println("Measurement MQ Transport running...")
		for {
			select {
			case msg := <-consume:
				var pmsg pipeline.Message
				if err := json.Unmarshal(msg.Body, &pmsg); err != nil {
					if nerr := msg.Nack(false, false); nerr != nil {
						err = fmt.Errorf("error nacking message: %w, while handling another error: %w", nerr, err)
					}
					log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
					continue
				}

				if err := svc.StorePipelineMessage(pmsg); err != nil {
					if nerr := msg.Nack(false, false); nerr != nil {
						err = fmt.Errorf("error nacking message: %w, while handling another error: %w", nerr, err)
					}
					log.Printf("Error storing pipeline message: %v\n", err)
					// Create error
					msgError := pipeline.PipelineError{
						ReceivedByWorker: pmsg,
						Error:            err.Error(),
						Timestamp:        time.Now().UnixMilli(),
						Worker:           "core-measurements",
					}
					msgErrorBytes, err := json.Marshal(msgError)
					if err != nil {
						log.Printf("error marshalling pipeline ErrorMessage into json: %v\n", err)
						continue
					}
					publish <- mq.PublishMessage{
						Topic: measurementErrorTopic,
						Publishing: amqp091.Publishing{
							Body: msgErrorBytes,
						},
					}

					continue
				}
				if err := msg.Ack(false); err != nil {
					log.Printf("Error Acking message: %s\n", err.Error())
				}
			case <-done:
				return
			}
		}
	}()

	return func() {
		close(done)
	}
}
