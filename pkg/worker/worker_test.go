package worker

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func TestWorkerSuite(t *testing.T) {
	suite.Run(t, new(workerSuite))
}

func (s *workerSuite) TestWorkerProcessorReturnsAnError() {
	// Arrange
	acker := &ackMock{}
	publisher := make(chan mq.PublishMessage)
	consumer := make(chan amqp091.Delivery)
	incomingMessage := pipeline.Message{
		ID: "blabla",
	}
	expectedMessage := pipeline.PipelineError{
		ReceivedByWorker: pipeline.Message{
			ID: "blabla",
		},
		Worker: "some-worker",
		Topic:  "message came from this topic",
		Error:  "unexpected error occurred!!",
	}
	w := worker{
		id:         "some-worker",
		mqErrTopic: "this is an error topic",
		mqQueue:    "message came from this topic",
		publisher:  publisher,
		consumer:   consumer,
		processor: func(m pipeline.Message) (pipeline.Message, error) {
			return pipeline.Message{}, fmt.Errorf("unexpected error occurred!!")
		},
		cancelToken: make(chan any),
	}

	// Act
	go w.Run()
	consumer <- amqp091.Delivery{
		Acknowledger: acker,
		Body:         toBytes(incomingMessage),
	}
	result := <-publisher
	close(consumer)
	<-w.cancelToken

	// Assert
	s.Equal(0, acker.ackCalled)
	s.Equal(1, acker.nackCalled)
	s.Equal(0, acker.rejectCalled)
	s.Equal(mq.PublishMessage{
		Topic: "this is an error topic",
		Publishing: amqp091.Publishing{
			Body: toBytes(expectedMessage),
		},
	}, result)
}

func (s *workerSuite) TestIncomingMessageIsInvalidJson() {
	// Arrange
	acker := &ackMock{}
	publisher := make(chan mq.PublishMessage)
	consumer := make(chan amqp091.Delivery)
	expectedMessage := pipeline.PipelineError{
		Worker:           "some-worker",
		ReceivedByWorker: pipeline.Message{},
		Topic:            "message came from this topic",
		Error:            "json: cannot unmarshal string into Go value of type pipeline.Message",
	}
	w := worker{
		id:          "some-worker",
		mqQueue:     "message came from this topic",
		mqErrTopic:  "this is an error topic",
		publisher:   publisher,
		consumer:    consumer,
		cancelToken: make(chan any),
	}

	// Act
	go w.Run()
	consumer <- amqp091.Delivery{
		Acknowledger: acker,
		Body:         toBytes("broken json!!"),
	}
	result := <-publisher
	close(consumer)
	<-w.cancelToken

	// Assert
	s.Equal(0, acker.ackCalled)
	s.Equal(1, acker.nackCalled)
	s.Equal(0, acker.rejectCalled)
	s.Equal(mq.PublishMessage{
		Topic: "this is an error topic",
		Publishing: amqp091.Publishing{
			Body: toBytes(expectedMessage),
		},
	}, result)
}

func (s *workerSuite) TestIncomingMessageWithNextStep() {
	// Arrange
	acker := &ackMock{}
	publisher := make(chan mq.PublishMessage)
	consumer := make(chan amqp091.Delivery)
	incomingMessage := pipeline.Message{
		ID:            "very-unique-id",
		StepIndex:     0,
		PipelineSteps: []string{"step1", "step2"},
		Measurements:  []pipeline.Measurement{},
	}
	expectedMessage := pipeline.Message{
		ID:            "very-unique-id",
		StepIndex:     1,
		PipelineSteps: []string{"step1", "step2"},
		Measurements:  []pipeline.Measurement{},
	}
	w := worker{
		publisher: publisher,
		consumer:  consumer,
		processor: func(m pipeline.Message) (pipeline.Message, error) {
			return m, nil
		},
		cancelToken: make(chan any),
	}

	// Act
	go w.Run()
	consumer <- amqp091.Delivery{
		Acknowledger: acker,
		Body:         toBytes(incomingMessage),
	}
	result := <-publisher
	close(consumer)
	<-w.cancelToken

	// Assert
	s.Equal(1, acker.ackCalled)
	s.Equal(0, acker.nackCalled)
	s.Equal(0, acker.rejectCalled)
	s.Equal(mq.PublishMessage{
		Topic: "step2",
		Publishing: amqp091.Publishing{
			MessageId: expectedMessage.ID,
			Body:      toBytes(expectedMessage),
		},
	}, result)
}

type workerSuite struct {
	suite.Suite
}

func toBytes[T interface{}](val T) []byte {
	b, err := json.Marshal(&val)
	if err != nil {
		panic(err)
	}
	return b
}

type ackMock struct {
	ackCalled    int
	nackCalled   int
	rejectCalled int
}

func (m *ackMock) Ack(tag uint64, multiple bool) error {
	m.ackCalled++
	return nil
}

func (m *ackMock) Nack(tag uint64, multiple bool, requeue bool) error {
	m.nackCalled++
	return nil
}

func (m *ackMock) Reject(tag uint64, requeue bool) error {
	m.rejectCalled++
	return nil
}
