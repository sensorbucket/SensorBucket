package pipeline_test

import (
	"math"
	"testing"
	"time"

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

func TestNewMessageRandomUUID(t *testing.T) {
	ids := []string{}
	for i := 10; i > 0; i-- {
		ids = append(ids, pipeline.NewMessage("", nil).ID)
	}
	for i, id := range ids {
		for j, id2 := range ids {
			if i == j {
				continue
			}
			assert.NotEqual(t, id, id2)
		}
	}
}

func TestNewMessageTimesInMillis(t *testing.T) {
	msg := pipeline.NewMessage("", nil)
	diff := float64(time.Now().UnixMilli() - msg.Timestamp)
	diff2 := float64(time.Now().UnixMilli() - msg.ReceivedAt)
	if math.Abs(diff) > 1000 || math.Abs(diff2) > 1000 {
		t.Errorf("Time between now and message timestamp/receivedAt differs more than expected: %.4f, it might not use millis!", diff)
	}
}

func TestHasReceivedDate(t *testing.T) {
	msg := pipeline.NewMessage("", nil)
	assert.GreaterOrEqual(t, time.Now().UnixMilli(), msg.ReceivedAt)
}
