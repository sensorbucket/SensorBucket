package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllStepsAddsCorrectRemainingSteps(t *testing.T) {
	// Arrange
	device := asPointer(int64(542))
	enriched := EnrichedSteps{
		EnrichedStep{
			Step: Step{
				TracingID:      "blabla",
				StepIndex:      0,
				StepsRemaining: 5,
				StartTime:      5432,
			},
			Status:                  Success,
			HighestCollectiveStatus: Failed,
		},
		EnrichedStep{
			Step: Step{
				TracingID:      "blabla",
				StepIndex:      1,
				StepsRemaining: 4,
				StartTime:      434234324,
				DeviceID:       device,
			},
			Status: Success,
		},
		EnrichedStep{
			Step: Step{
				TracingID:      "blabla",
				StepIndex:      2,
				StepsRemaining: 3,
				StartTime:      3253254354,
				Error:          asPointer("some weird error occurred!!"),
			},
			Status: Failed,
		},
	}

	// Act
	all := enriched.AllSteps()

	// Assert
	assert.Equal(t, Failed, all.TotalStatus())
	assert.Equal(t, int64(5432), all.TotalStartTime())
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
