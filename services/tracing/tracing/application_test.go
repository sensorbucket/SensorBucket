package tracing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func TestPipelineMessageWithInvalidAmountOfPipelineSteps(t *testing.T) {
	svc := Service{}
	assert.EqualError(t, svc.HandlePipelineMessage(
		pipeline.Message{
			PipelineSteps: []string{"1", "2", "3"},
			StepIndex:     3,
		},
		time.Now()), "steps remaining cannot be smaller than 0 (pipelinesteps len: 3, stepindex: 3)")
	assert.EqualError(t, svc.HandlePipelineError(
		pipeline.PipelineError{
			ReceivedByWorker: pipeline.Message{
				PipelineSteps: []string{"1", "2", "3", "4", "5"},
				StepIndex:     5,
			},
		},
		time.Now()), "steps remaining cannot be smaller than 0 (pipelinesteps len: 5, stepindex: 5)")
}

func TestPipelineErrorAppears(t *testing.T) {
	// Arrange
	type scene struct {
		input    pipeline.PipelineError
		expected Step
	}
	publishTime := time.Now()
	e := asPointer("some weird error occurred!!")
	scenarios := map[string]scene{
		"pipeline error with 3 steps remaining": {
			input: pipeline.PipelineError{
				ReceivedByWorker: pipeline.Message{
					TracingID:            "234324",
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
				StartTime:      publishTime,
				Error:          e,
			},
		},
		"pipeline message with 0 steps remaining": {
			input: pipeline.PipelineError{
				ReceivedByWorker: pipeline.Message{
					TracingID:            "234324",
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
				StartTime:      publishTime,
				Error:          e,
			},
		},
		"pipeline message with 1 step remaining": {
			input: pipeline.PipelineError{
				ReceivedByWorker: pipeline.Message{
					TracingID:            "234324",
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
				StartTime:      publishTime,
				Error:          e,
			},
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			stepStore := StepStoreMock{
				UpsertStepFunc: func(step Step, withError bool) error {
					return nil
				},
			}
			tracingService := Service{
				stepStore: &stepStore,
			}

			// Act and Assert
			assert.NoError(t, tracingService.HandlePipelineError(cfg.input, publishTime))
			assert.Equal(t, []struct {
				Step      Step
				WithError bool
			}{
				{
					Step:      cfg.expected,
					WithError: true,
				},
			}, stepStore.UpsertStepCalls())
		})
	}
}

func TestPipelineMessageAppears(t *testing.T) {
	// Arrange
	type scene struct {
		input    pipeline.Message
		expected Step
	}
	publishTime := time.Now()
	scenarios := map[string]scene{
		"pipeline message with 3 steps remaining": {
			input: pipeline.Message{
				TracingID:            "234324",
				Timestamp:     21342143,
				StepIndex:     3,
				PipelineSteps: []string{"A", "B", "C", "D", "E", "F", "G"},
			},
			expected: Step{
				TracingID:      "234324",
				StepIndex:      3,
				StepsRemaining: 3,
				StartTime:      publishTime,
			},
		},
		"pipeline message with 0 steps remaining": {
			input: pipeline.Message{
				TracingID:            "234324",
				Timestamp:     21342143,
				StepIndex:     3,
				PipelineSteps: []string{"A", "B", "C", "D"},
			},
			expected: Step{
				TracingID:      "234324",
				StepIndex:      3,
				StepsRemaining: 0,
				StartTime:      publishTime,
			},
		},
		"pipeline message with 1 step remaining": {
			input: pipeline.Message{
				TracingID:            "234324",
				Timestamp:     21342143,
				StepIndex:     4,
				PipelineSteps: []string{"A", "B", "C", "D", "E", "F"},
			},
			expected: Step{
				TracingID:      "234324",
				StepIndex:      4,
				StepsRemaining: 1,
				StartTime:      publishTime,
			},
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			stepStore := StepStoreMock{
				UpsertStepFunc: func(step Step, withError bool) error {
					return nil
				},
			}
			tracingService := Service{
				stepStore: &stepStore,
			}

			// Act and Assert
			assert.NoError(t, tracingService.HandlePipelineMessage(cfg.input, publishTime))
			assert.Equal(t, []struct {
				Step      Step
				WithError bool
			}{
				{
					Step:      cfg.expected,
					WithError: false,
				},
			}, stepStore.UpsertStepCalls())
		})
	}
}
