package tracing

import (
	"time"

	"github.com/samber/lo"
)

type Step struct {
	TracingID      string `pg:"tracing_id"`
	StepIndex      int64  `pg:"step_index"`
	StepsRemaining int64  `pg:"steps_remaining"`
	StartTime      int64  `pg:"start_time"`
	Error          string `pg:"error"`
}

// EnrichedStep contains extra properties derived from the data stored in the database using the Step model
type EnrichedStep struct {
	Step
	TotalStatus Status
	Status      Status
	Duration    time.Duration
}

type EnrichedSteps []EnrichedStep

func (es EnrichedSteps) TotalStatus() Status {
	// The total status is always the highest status in the step list
	return lo.MaxBy(es, func(item, max EnrichedStep) bool {
		return item.Status > max.Status
	}).Status
}

func (es EnrichedSteps) AllSteps() []EnrichedStep {
	if len(es) == 0 {
		return []EnrichedStep{}
	}

	// Enriched Steps are only models from the database, some steps do not exist yet in the database, however we do
	// want to include them in the all steps list. Using the status from the last step we can derive the state of all remaining steps

	// The last step status and steps remaining determines if any steps need to be added to the step list
	lastStep := es[len(es)-1]
	remainingStatus := Pending
	if lastStep.Status != Failed {
		// When the last step in the list has failed all the remaining steps are canceled
		remainingStatus = Canceled
	}

	// Convert the available enriched steps to the
	steps := lo.Times(len(es)+int(lastStep.StepsRemaining), func(index int) EnrichedStep {
		if index >= len(es) {
			// We are past the last available step in the step list
			// add any remaining (non-existent in the database) steps and derive
			// their states from the lastStep in the list
			return EnrichedStep{
				Step: Step{
					TracingID:      lastStep.TracingID,
					StepIndex:      int64(index),
					StepsRemaining: int64(len(es) + int(lastStep.StepsRemaining) - index + 1),
				},
				TotalStatus: lastStep.TotalStatus,
				Status:      remainingStatus,
			}
		}
		return es[index]
	})

	return steps
}

type Status int

const (
	Unknown    Status = 0
	Canceled   Status = 1
	Pending    Status = 2
	Success    Status = 3
	InProgress Status = 4
	Failed     Status = 5
)

func (s Status) String() string {
	switch s {
	case Canceled:
		return "canceled"
	case Pending:
		return "pending"
	case Success:
		return "success"
	case InProgress:
		return "in progress"
	case Failed:
		return "failed"
	}
	return "unknown"
}
