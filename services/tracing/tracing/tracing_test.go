package tracing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnrichedStepsProperties(t *testing.T) {
	// Arrange
	device := int64(542)
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

func TestAllStepsAppendsMissingSteps(t *testing.T) {
	type scenario struct {
		input    EnrichedSteps
		expected EnrichedSteps
	}

	// Arrange
	scenarios := map[string]scenario{
		"3 steps but 1 step is missing": {
			input: sl(
				s(0, 2, Success),
				s(2, 0, Success),
			),
			expected: sl(
				s(0, 2, Success),
				s(1, 1, Success),
				s(2, 0, Success),
			),
		},
		"7 steps but multiple steps are missing": {
			input: sl(
				s(0, 6, Success),
				s(2, 4, Success),
				s(4, 2, Success),
				s(6, 0, Success),
			),
			expected: sl(
				s(0, 6, Success),
				s(1, 5, Success),
				s(2, 4, Success),
				s(3, 3, Success),
				s(4, 2, Success),
				s(5, 1, Success),
				s(6, 0, Success),
			),
		},
		"first steps are missing, last step is pending": {
			input: sl(
				s(2, 3, Success),
				s(3, 2, Success),
				s(4, 1, InProgress),
			),
			expected: sl(
				s(0, 5, Success),
				s(1, 4, Success),
				s(2, 3, Success),
				s(3, 2, Success),
				s(4, 1, InProgress),
				s(5, 0, Pending),
			),
		},
		"multiple steps pending": {
			input: sl(
				s(0, 4, Success),
			),
			expected: sl(
				s(0, 4, Success),
				s(1, 3, Pending),
				s(2, 2, Pending),
				s(3, 1, Pending),
				s(4, 0, Pending),
			),
		},
		"all steps success": {
			input: sl(
				s(0, 4, Success),
				s(1, 3, Success),
				s(2, 2, Success),
				s(3, 1, Success),
				s(4, 0, Success),
			),
			expected: sl(
				s(0, 4, Success),
				s(1, 3, Success),
				s(2, 2, Success),
				s(3, 1, Success),
				s(4, 0, Success),
			),
		},
		"no steps": {
			input:    EnrichedSteps{},
			expected: EnrichedSteps{},
		},
		"some failed and canceled steps": {
			input: sl(
				s(0, 4, Success),
				s(1, 3, Failed),
			),
			expected: sl(
				s(0, 4, Success),
				s(1, 3, Failed),
				s(2, 2, Canceled),
				s(3, 1, Canceled),
				s(4, 0, Canceled),
			),
		},
		"multiple missing steps": {
			input: sl(
				s(9, 0, Success),
			),
			expected: sl(
				s(0, 9, Success),
				s(1, 8, Success),
				s(2, 7, Success),
				s(3, 6, Success),
				s(4, 5, Success),
				s(5, 4, Success),
				s(6, 3, Success),
				s(7, 2, Success),
				s(8, 1, Success),
				s(9, 0, Success)),
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			// Act and Assert
			assert.Equal(t, cfg.expected, cfg.input.AllSteps())
		})
	}
}

func asPointer[T any](val T) *T {
	return &val
}

func s(stepIndex uint, stepsRemaining uint, status Status) EnrichedStep {
	return EnrichedStep{
		Step: Step{
			TracingID:      "blablabla",
			StepIndex:      uint64(stepIndex),
			StepsRemaining: uint64(stepsRemaining),
		},
		Status: status,
	}
}

func sl(enr ...EnrichedStep) EnrichedSteps {
	steps := EnrichedSteps{}
	for _, s := range enr {
		steps = append(steps, s)
	}
	return steps
}
