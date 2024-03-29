package tracing

import (
	"time"

	"github.com/samber/lo"
)

type Status int

const (
	Unknown Status = iota
	Canceled
	Pending
	Success
	InProgress
	Failed
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
	default:
		return "unknown"
	}
}

type Step struct {
	TracingID      string
	StepIndex      uint64
	StepsRemaining uint64
	StartTime      time.Time
	DeviceID       int64
	Error          *string
}

// EnrichedStep contains extra properties derived from the data stored in the database using the Step model
type EnrichedStep struct {
	Step
	HighestCollectiveStatus Status
	Status                  Status
	Duration                time.Duration
}

type EnrichedSteps []EnrichedStep

func (es EnrichedSteps) TotalStartTime() time.Time {
	return lo.MinBy(es, func(item, min EnrichedStep) bool {
		// A StartTime with value 0 is considered to be not set
		return item.StartTime.UnixMilli() > 0 && item.StartTime.UnixMilli() < min.StartTime.UnixMilli()
	}).StartTime
}

func (es EnrichedSteps) TotalStatus() Status {
	if len(es) == 0 {
		return Unknown
	}
	return es[0].HighestCollectiveStatus
}

func (es EnrichedSteps) DeviceID() int64 {
	return lo.MaxBy(es, func(a, b EnrichedStep) bool {
		return a.DeviceID > b.DeviceID
	}).DeviceID
}

func (es EnrichedSteps) Exists(stepIndex uint64) bool {
	_, ok := lo.Find(es, func(item EnrichedStep) bool {
		return stepIndex == item.StepIndex
	})
	return ok
}

func (es EnrichedSteps) IsLastStep(step EnrichedStep) bool {
	return es[len(es)-1] == step
}

func (es EnrichedSteps) AllSteps() EnrichedSteps {
	if len(es) == 0 {
		return []EnrichedStep{}
	}

	// Enriched Steps are only models from the database, some steps do not exist yet in the database, however we do
	// want to include them in the all steps list. Using the status from the last step we can derive the state of all remaining steps

	// The last step status and steps remaining determines if any steps need to be added to the step list
	lastStep := es[len(es)-1]

	remainingStatus := Pending
	if lastStep.Status == Failed {
		// When the last step in the list has failed all the remaining steps are canceled
		remainingStatus = Canceled
	}

	totalSteps := int(lastStep.StepsRemaining + lastStep.StepIndex + 1)
	indexOffset := 0

	steps := lo.Times(int(lastStep.StepIndex+1)+int(lastStep.StepsRemaining), func(index int) (enrStep EnrichedStep) {
		if index > int(lastStep.StepIndex) {
			// We are past the last available step in the step list
			// add any remaining (non-existent in the database) steps and derive
			// their states from the lastStep in the list
			return EnrichedStep{
				Step: Step{
					TracingID:      lastStep.TracingID,
					StepIndex:      uint64(index),
					StepsRemaining: uint64(totalSteps - index - 1),
				},
				Status: remainingStatus,
			}
		}

		if !es.Exists(uint64(index)) {
			// There is a step missing somewhere in between the steps in the list.
			indexOffset++
			return EnrichedStep{
				Step: Step{
					TracingID:      lastStep.TracingID,
					StepIndex:      uint64(index),
					StepsRemaining: uint64(totalSteps - index - 1),
				},

				// if a step is missing in between the other steps, it's safe to assume the step has been succesfully executed
				// otherwise the next steps would've been canceled.
				Status: Success,
			}
		}

		// if the next step doesn't exist but there is a next step in the list we have to adjust the status of this step
		// if any steps exists after this step the status has to be succesful otherwise the next steps would've been canceled
		step := es[index-indexOffset]
		if step.Status == InProgress && !es.IsLastStep(step) {
			step.Status = Success
		}
		return step
	})

	return steps
}

func StatusStringsToStatusCodes(statusses []string) []int64 {
	codes := []int64{}
	for _, s := range statusses {
		for _, c := range allPossibleCodes {
			if c.String() == s {
				codes = append(codes, int64(c))
			}
		}
	}
	return codes
}

var allPossibleCodes = []Status{
	Unknown,
	Canceled,
	Pending,
	Success,
	InProgress,
	Failed,
}
