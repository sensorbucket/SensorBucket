package tracing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func TestTracingSuite(t *testing.T) {
	suite.Run(t, new(tracingSuite))
}

func (s *tracingSuite) TestPipelineMessageAppears() {
	// Arrange
	startTime := time.Date(2023, time.August, 21, 12, 30, 0, 0, time.UTC)
	pipelineMessage := pipeline.Message{
		ID:            "some-random-id",
		StepIndex:     3, // The next step to be executed
		Timestamp:     startTime.Unix(),
		PipelineSteps: []string{"A", "B", "C", "D"},
	}
	expected := Step{
		TracingID:      "some-random-id",
		StepIndex:      2, // Actual completed step
		StepsRemaining: 1,
		StartTime:      startTime.Unix(),
		Error:          "",
	}
	stepStore := stepStoreMock{}
	pipelineMessages := make(chan pipeline.Message)
	errorMessage := make(chan PipelineError)
	tracingService := Service{
		stepStore:        &stepStore,
		pipelineMessages: pipelineMessages,
		errorMessages:    errorMessage,
	}

	// Act
	go tracingService.Run()
	pipelineMessages <- pipelineMessage
	close(errorMessage)
	close(pipelineMessages)

	// Assert
	s.Equal(expected, stepStore.inserted)
}

type tracingSuite struct {
	suite.Suite
}

type stepStoreMock struct {
	inserted Step
}

func (s *stepStoreMock) AddStep(step Step) error {
	s.inserted = step
	return nil
}
