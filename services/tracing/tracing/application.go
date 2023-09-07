package tracing

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type StepStore interface {
	UpsertStep(step Step, withError bool) error
	GetStepsByTracingIDs([]string) ([]EnrichedStep, error)
	QueryTraces(filter Filter, r pagination.Request) (*pagination.Page[string], error)
}

func New(stepStore StepStore) *Service {
	return &Service{
		stepStore: stepStore,
	}
}

type Service struct {
	stepStore StepStore
}

func (s *Service) HandlePipelineMessage(pipelineMessage pipeline.Message, time time.Time) error {
	if len(pipelineMessage.PipelineSteps)-(int(pipelineMessage.StepIndex+1)) < 0 {
		return fmt.Errorf("steps remaining cannot be smaller than 0 (pipelinesteps len: %d, stepindex: %d)",
			len(pipelineMessage.PipelineSteps),
			pipelineMessage.StepIndex,
		)
	}

	step := Step{
		TracingID: pipelineMessage.ID,
		StepIndex: pipelineMessage.StepIndex,

		// We have to add 1 to the stepindex to get the actual steps remaining
		StepsRemaining: uint64(len(pipelineMessage.PipelineSteps) - (int(pipelineMessage.StepIndex + 1))),

		// The timestamp is set by the mq when it is send to the queue. The next step's starttime can be used to deduce the processing time between the 2 steps
		// this duration consists of: Time in Queue and the Processing Time in the worker
		StartTime: time,
	}

	if pipelineMessage.Device != nil {
		step.DeviceID = &pipelineMessage.Device.ID
	}

	return s.stepStore.UpsertStep(step, false)
}

func (s *Service) HandlePipelineError(errorMessage pipeline.PipelineError, time time.Time) error {
	if len(errorMessage.ReceivedByWorker.PipelineSteps)-(int(errorMessage.ReceivedByWorker.StepIndex+1)) < 0 {
		return fmt.Errorf("steps remaining cannot be smaller than 0 (pipelinesteps len: %d, stepindex: %d)",
			len(errorMessage.ReceivedByWorker.PipelineSteps),
			errorMessage.ReceivedByWorker.StepIndex,
		)
	}

	step := Step{
		TracingID:      errorMessage.ReceivedByWorker.ID,
		StepIndex:      errorMessage.ReceivedByWorker.StepIndex,
		StepsRemaining: uint64(len(errorMessage.ReceivedByWorker.PipelineSteps) - (int(errorMessage.ReceivedByWorker.StepIndex + 1))),
		StartTime:      time,
		Error:          &errorMessage.Error,
	}

	if errorMessage.ReceivedByWorker.Device != nil {
		step.DeviceID = &errorMessage.ReceivedByWorker.Device.ID
	}

	if errorMessage.ProcessingAttempt.Device != nil {
		step.DeviceID = &errorMessage.ProcessingAttempt.Device.ID
	}

	return s.stepStore.UpsertStep(step, true)
}

func (s *Service) QueryTraces(f Filter, r pagination.Request) (*pagination.Page[TraceDTO], error) {
	// Retrieve all the traces according to it's pagination first
	filteredTraces, err := s.stepStore.QueryTraces(f, r)
	if err != nil {
		return nil, err
	}

	// TODO: this is not a maintainable solution, the second query might receive a thousand values in the 'IN' clause

	// Now enrich the trace data with the step data
	steps, err := s.stepStore.GetStepsByTracingIDs(filteredTraces.Data)
	if err != nil {
		return nil, err
	}

	// Now change the map the filtered trace ids to the correct steps and keep the order of traces
	traces := []TraceDTO{}
	lo.ForEach(filteredTraces.Data, func(tracingId string, index int) {
		enrichedSteps := EnrichedSteps(lo.Filter(steps, func(s EnrichedStep, index int) bool {
			return s.TracingID == tracingId
		}))
		traces = append(traces, TraceDTO{
			TracingId: tracingId,
			Status:    enrichedSteps.TotalStatus().String(),
			StartTime: enrichedSteps.TotalStartTime(),
			DeviceId:  enrichedSteps.DeviceID(),

			// The EnrichedSteps also have to be updated to a DTO object
			Steps: lo.Map(enrichedSteps.AllSteps(), func(val EnrichedStep, _ int) StepDTO {
				stepDto := StepDTO{
					Status:   val.Status.String(),
					Duration: val.Duration,
				}
				if val.Error != nil {
					stepDto.Error = *val.Error
				}
				return stepDto
			}),
		})
	})

	return &pagination.Page[TraceDTO]{
		Cursor: filteredTraces.Cursor,
		Data:   traces,
	}, nil
}

type Filter struct {
	TracingIds          []string `schema:"tracing_id"`
	Status              []string `schema:"status"`
	DeviceIds           []int64  `schema:"device_id"`
	DurationGreaterThan *int64   `schema:"duration_greater_than"`
	DurationLowerThan   *int64   `schema:"duration_lower_than"`
}

type TraceDTO struct {
	TracingId string    `json:"tracing_id"`
	DeviceId  *int64    `json:"device_id"`
	StartTime time.Time `json:"start_time"`
	Status    string    `json:"status"`
	Steps     []StepDTO `json:"steps"`
}

type StepDTO struct {
	Status   string         `json:"status"`
	Duration *time.Duration `json:"duration"`
	Error    string         `json:"error"`
}
