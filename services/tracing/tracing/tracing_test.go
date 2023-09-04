package tracing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAllStepsAddsCorrectRemainingSteps(t *testing.T) {
	// Arrange
	device := asPointer(int64(542))
	startT := time.Now()
	enriched := EnrichedSteps{
		EnrichedStep{
			Step: Step{
				TracingID:      "blabla",
				StepIndex:      0,
				StepsRemaining: 5,
				StartTime:      startT,
			},
			Status:                  Success,
			HighestCollectiveStatus: Failed,
		},
		EnrichedStep{
			Step: Step{
				TracingID:      "blabla",
				StepIndex:      1,
				StepsRemaining: 4,
				StartTime:      startT.Add(time.Second * 10),
				DeviceID:       device,
			},
			Status: Success,
		},
		EnrichedStep{
			Step: Step{
				TracingID:      "blabla",
				StepIndex:      2,
				StepsRemaining: 3,
				StartTime:      startT.Add(time.Second * 10),
				Error:          asPointer("some weird error occurred!!"),
			},
			Status: Failed,
		},
	}

	// Act
	all := enriched.AllSteps()

	// Assert
	assert.Equal(t, Failed, all.TotalStatus())
	assert.Equal(t, startT, all.TotalStartTime())
	assert.Equal(t, device, all.DeviceID())
	assert.Equal(t,
		append(enriched,
			EnrichedStep{
				Step: Step{
					TracingID:      "blabla",
					StepIndex:      3,
					StepsRemaining: 2,
				},
				Status: Canceled,
			},
			EnrichedStep{
				Step: Step{
					TracingID:      "blabla",
					StepIndex:      4,
					StepsRemaining: 1,
				},
				Status: Canceled,
			},
			EnrichedStep{
				Step: Step{
					TracingID:      "blabla",
					StepIndex:      5,
					StepsRemaining: 0,
				},
				Status: Canceled,
			}),
		all)
}

func asPointer[T any](val T) *T {
	return &val
}
