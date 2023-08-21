package pipeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func TestMessageNextStepErrorOnEmptyPipelineSteps(t *testing.T) {
	msg := pipeline.Message{
		PipelineSteps: []string{""},
	}
	next, err := msg.NextStep()
	assert.Equal(t, next, "")
	assert.Error(t, err)
}

func TestMessageNextStepErrorOnLastStep(t *testing.T) {
	msg := pipeline.Message{
		StepIndex:     2,
		PipelineSteps: []string{"a", "b", "c"},
	}
	next, err := msg.NextStep()
	assert.Equal(t, next, "")
	assert.Error(t, err)
}

func TestMessageNextStepFewRemainingSteps(t *testing.T) {
	msg := pipeline.Message{
		StepIndex:     1,
		PipelineSteps: []string{"a", "b", "c"},
	}
	next, err := msg.NextStep()
	assert.Equal(t, next, "c")
	assert.NoError(t, err)
}
