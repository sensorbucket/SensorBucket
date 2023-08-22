package tracing

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func TestTracingSuite(t *testing.T) {
	suite.Run(t, new(tracingSuite))
}

func (s *tracingSuite) TestPipelineErrorAppears() {
	// Arrange
	stepStore := stepStoreMock{}
	tracingService := Service{
		stepStore: &stepStore,
	}

	type scene struct {
		input    pipeline.PipelineError
		expected Step
	}
	scenarios := map[string]scene{
		"pipeline error with 3 steps remaining": {
			input: pipeline.PipelineError{
				ReceivedByWorker: pipeline.Message{
					ID:            "234324",
					Timestamp:     21342143,
					StepIndex:     3,
					PipelineSteps: []string{"A", "B", "C", "D", "E", "F", "G"},
				},
				Error: "some weird error occurred!!",
			},
			expected: Step{
				TracingID:      "234324",
				StepIndex:      3,
				StepsRemaining: 3,
				StartTime:      21342143,
				Error:          "some weird error occurred!!",
			},
		},
		"pipeline message with 0 steps remaining": {
			input: pipeline.PipelineError{
				ReceivedByWorker: pipeline.Message{
					ID:            "234324",
					Timestamp:     21342143,
					StepIndex:     6,
					PipelineSteps: []string{"A", "B", "C", "D", "E", "F", "G"},
				},
				Error: "some weird error occurred!!",
			},
			expected: Step{
				TracingID:      "234324",
				StepIndex:      6,
				StepsRemaining: 0,
				StartTime:      21342143,
				Error:          "some weird error occurred!!",
			},
		},
		"pipeline message with 1 step remaining": {
			input: pipeline.PipelineError{
				ReceivedByWorker: pipeline.Message{
					ID:            "234324",
					Timestamp:     21342143,
					StepIndex:     5,
					PipelineSteps: []string{"A", "B", "C", "D", "E", "F", "G"},
				},
				Error: "some weird error occurred!!",
			},
			expected: Step{
				TracingID:      "234324",
				StepIndex:      5,
				StepsRemaining: 1,
				StartTime:      21342143,
				Error:          "some weird error occurred!!",
			},
		},
	}

	for scene, cfg := range scenarios {
		s.Run(scene, func() {
			// Act and Assert
			s.NoError(tracingService.HandlePipelineError(cfg.input))
			s.Equal(cfg.expected, stepStore.inserted)
		})
	}
}

func (s *tracingSuite) TestPipelineMessageAppears() {
	// Arrange
	stepStore := stepStoreMock{}
	tracingService := Service{
		stepStore: &stepStore,
	}

	type scene struct {
		input    pipeline.Message
		expected Step
	}
	scenarios := map[string]scene{
		"pipeline message with 3 steps remaining": {
			input: pipeline.Message{
				ID:            "234324",
				Timestamp:     21342143,
				StepIndex:     3,
				PipelineSteps: []string{"A", "B", "C", "D", "E", "F", "G"},
			},
			expected: Step{
				TracingID:      "234324",
				StepIndex:      3,
				StepsRemaining: 3,
				StartTime:      21342143,
				Error:          "",
			},
		},
		"pipeline message with 0 steps remaining": {
			input: pipeline.Message{
				ID:            "234324",
				Timestamp:     21342143,
				StepIndex:     3,
				PipelineSteps: []string{"A", "B", "C", "D"},
			},
			expected: Step{
				TracingID:      "234324",
				StepIndex:      3,
				StepsRemaining: 0,
				StartTime:      21342143,
				Error:          "",
			},
		},
		"pipeline message with 1 step remaining": {
			input: pipeline.Message{
				ID:            "234324",
				Timestamp:     21342143,
				StepIndex:     4,
				PipelineSteps: []string{"A", "B", "C", "D", "E", "F"},
			},
			expected: Step{
				TracingID:      "234324",
				StepIndex:      4,
				StepsRemaining: 1,
				StartTime:      21342143,
				Error:          "",
			},
		},
	}

	for scene, cfg := range scenarios {
		s.Run(scene, func() {
			// Act and Assert
			s.NoError(tracingService.HandlePipelineMessage(cfg.input))
			s.Equal(cfg.expected, stepStore.inserted)
		})
	}
}

type tracingSuite struct {
	suite.Suite
}

type stepStoreMock struct {
	inserted Step
}

func (s *stepStoreMock) Insert(step Step) error {
	s.inserted = step
	return nil
}
