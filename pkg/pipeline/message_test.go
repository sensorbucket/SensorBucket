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

func TestExpectCurrentStep(t *testing.T) {
	testCases := []struct {
		desc     string
		ix       int
		expected string
		err      error
	}{
		{
			desc:     "",
			ix:       0,
			expected: "a",
		},
		{
			desc:     "",
			ix:       1,
			expected: "b",
		},
		{
			desc:     "",
			ix:       3,
			expected: "d",
		},
		{
			desc:     "",
			ix:       55,
			expected: "",
			err:      pipeline.ErrMessageNoSteps,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			msg := pipeline.Message{
				PipelineSteps: []string{"a", "b", "c", "d"},
				StepIndex:     uint64(tC.ix),
			}
			step, err := msg.CurrentStep()
			if tC.err != nil {
				assert.Error(t, tC.err, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tC.expected, step)
			}
		})
	}
}
