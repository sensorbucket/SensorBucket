package tracing

import (
	"time"

	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type StepStore interface {
	Insert(Step) error
	Query(Filter, pagination.Request) (*pagination.Page[EnrichedStep], error)
}

func New(stepStore StepStore) *Service {
	return &Service{
		stepStore: stepStore,
	}
}

type Service struct {
	stepStore StepStore
}

func (s *Service) HandlePipelineMessage(pipelineMessage pipeline.Message) error {
	step := Step{
		TracingID: pipelineMessage.ID,
		StepIndex: pipelineMessage.StepIndex,

		// We have to add 1 to the stepindex to get the actual steps remaining
		StepsRemaining: int64(len(pipelineMessage.PipelineSteps) - (int(pipelineMessage.StepIndex + 1))),

		// The timestamp is set by the mq when it is send to the queue. The next step's starttime can be used to deduce the processing time between the 2 steps
		// this duration consists of: Time in Queue and the Processing Time in the worker
		StartTime: pipelineMessage.Timestamp,
	}

	if pipelineMessage.Device != nil {
		step.DeviceId = &pipelineMessage.Device.ID
	}

	return s.addStep(step)
}

func (s *Service) HandlePipelineError(errorMessage pipeline.PipelineError) error {
	step := Step{
		TracingID:      errorMessage.ReceivedByWorker.ID,
		StepIndex:      errorMessage.ReceivedByWorker.StepIndex,
		StepsRemaining: int64(len(errorMessage.ReceivedByWorker.PipelineSteps) - (int(errorMessage.ReceivedByWorker.StepIndex + 1))),
		StartTime:      errorMessage.ReceivedByWorker.Timestamp,
		Error:          &errorMessage.Error,
	}

	if errorMessage.ReceivedByWorker.Device != nil {
		step.DeviceId = &errorMessage.ReceivedByWorker.Device.ID
	}

	if errorMessage.ProcessingAttempt.Device != nil {
		step.DeviceId = &errorMessage.ProcessingAttempt.Device.ID
	}

	return s.addStep(step)
}

func (s *Service) QueryTraces(f Filter, r pagination.Request) (*pagination.Page[TraceDTO], error) {
	page, err := s.stepStore.Query(f, r)
	if err != nil {
		return nil, err
	}

	// Group all the enriched steps to their corresponding trace id so we can build the correct TraceDTO
	grouped := lo.GroupBy(page.Data, func(step EnrichedStep) string {
		return step.TracingID
	})

	// Now change the map of trace ids with their steps to the required trace dto
	traces := lo.MapToSlice(grouped, func(key string, value []EnrichedStep) TraceDTO {

		asEnriched := EnrichedSteps(value)

		return TraceDTO{
			TracingId: key,
			Status:    asEnriched.TotalStatus().String(),
			StartTime: asEnriched.TotalStartTime(),
			DeviceId:  asEnriched.DeviceId(),

			// The EnrichedSteps also have to be updated to a DTO object
			Steps: lo.Map(asEnriched.AllSteps(), func(val EnrichedStep, _ int) StepDTO {
				stepDto := StepDTO{
					Status:   val.Status.String(),
					Duration: val.Duration,
				}
				if val.Error != nil {
					stepDto.Error = *val.Error
				}
				return stepDto
			}),
		}
	})

	return &pagination.Page[TraceDTO]{
		Cursor: page.Cursor,
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
	StartTime int64     `json:"start_time"`
	Status    string    `json:"status"`
	Steps     []StepDTO `json:"steps"`
}

type StepDTO struct {
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error"`
}

func (s *Service) addStep(step Step) error {
	return s.stepStore.Insert(step)
}
