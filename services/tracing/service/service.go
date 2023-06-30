package tracing

//go:generate moq -pkg tracing_test -out mock_test.go . MessageStateStorer MessageArchiver MessageStateIterator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func Run(ctx context.Context, amqpHost, amqpQueue string, state MessageStateStorer, archiver MessageArchiver) error {
	svc := New(state, archiver)
	amqp := mq.NewAMQPConsumer(amqpHost, amqpQueue, func(c *amqp091.Channel) error {
		// TODO: Setup queue, exchange etc
		_, err := c.QueueDeclare(amqpQueue, false, false, false, false, nil)
		return err
	})
	go amqp.Start()

	go func() {
		for del := range amqp.Consume() {
			svc.ProcessDelivery(del)
		}
	}()
	fmt.Print("Running\n")

	// Wait for interrupt
	<-ctx.Done()
	fmt.Print("Shutting down gracefully\n")
	amqp.Shutdown()
	return nil
}

type MessageStateStorer interface {
	UpdateState(ctx context.Context, id string, timestamp time.Time) error
	FinishState(ctx context.Context, id string) error
}

type MessageArchiver interface {
	Archive(ctx context.Context, delivery amqp091.Delivery) error
}

type Service struct {
	stateStore MessageStateStorer
	archiver   MessageArchiver
}

func New(stateStore MessageStateStorer, archiver MessageArchiver) *Service {
	return &Service{
		stateStore: stateStore,
		archiver:   archiver,
	}
}

func (s *Service) ProcessDelivery(del amqp091.Delivery) error {
	var err error
	var msg pipeline.Message
	if err := json.Unmarshal(del.Body, &msg); err != nil {
		fmt.Printf("Failed unmarshal incoming message to pipeline.message: %v\n", err)
		return errors.New("Unmarshal error")
	}

	if del.MessageId == "" {
		del.MessageId = uuid.NewString()
		fmt.Printf("Warning, delivery does not have MessageID set, using '%s' for archiving purposes\n", del.MessageId)
	}

	err = s.archiver.Archive(context.TODO(), del)
	if err != nil {
		fmt.Printf("Failed to archive message: %v\n", err)
		// Continue as its not critical
	}
	err = s.ProcessPipelineMessage(del.MessageId, del.RoutingKey, msg)
	if err != nil {
		fmt.Printf("Failed to process message: %v\n", err)
		del.Nack(false, false)
		return err
	}
	del.Ack(false)
	return nil
}

func (s *Service) ProcessPipelineMessage(id, topic string, msg pipeline.Message) error {
	ctx := context.TODO()

	err := s.stateStore.UpdateState(ctx, id, time.Now())
	if err != nil {
		return err
	}

	if len(msg.PipelineSteps) == 0 {
		// Finished
		err = s.stateStore.FinishState(ctx, id)
		if err != nil {
			return err
		}
	}

	return nil
}
