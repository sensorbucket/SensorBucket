package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func TestPipelineErrorAppears(t *testing.T) {
	// Arrange
	stepStore := stepStoreMock{}
	tracingService := Service{
		stepStore: &stepStore,
	}

	type scene struct {
		input    pipeline.PipelineError
		expected Step
	}
	e := asPointer("some weird error occurred!!")
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
				Error:          e,
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
				Error:          e,
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
				Error:          e,
			},
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			// Act and Assert
			assert.NoError(t, tracingService.HandlePipelineError(cfg.input))
			assert.Equal(t, cfg.expected, stepStore.inserted)
		})
	}
}

func TestPipelineMessageAppears(t *testing.T) {
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
			},
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			// Act and Assert
			assert.NoError(t, tracingService.HandlePipelineMessage(cfg.input))
			assert.Equal(t, cfg.expected, stepStore.inserted)
		})
	}
}

type stepStoreMock struct {
	inserted Step
}

func (s *stepStoreMock) Insert(step Step) error {
	s.inserted = step
	return nil
}

func (s *stepStoreMock) Query(Filter, pagination.Request) (*pagination.Page[EnrichedStep], error) {
	return &pagination.Page[EnrichedStep]{}, nil
}
