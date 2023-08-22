package tracing

import (
	"time"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type Status string

const (
	Unknown    Status = "unknown"
	Pending    Status = "pending"
	InProgress Status = "in progress"
	Success    Status = "success"
	Failed     Status = "failed"
	Canceled   Status = "canceled"
)

type StepStore interface {
	Insert(Step) error
	Query(Filter, pagination.Request) (*pagination.Page[Step], error)
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
	return s.addStep(Step{
		TracingID: pipelineMessage.ID,
		StepIndex: pipelineMessage.StepIndex,

		// We have to add 1 to the stepindex to get the actual steps remaining
		StepsRemaining: int64(len(pipelineMessage.PipelineSteps) - (int(pipelineMessage.StepIndex + 1))),

		// The timestamp is set by the mq when it is send to the queue. The next step's starttime can be used to deduce the processing time between the 2 steps
		StartTime: pipelineMessage.Timestamp,
	})
}

func (s *Service) HandlePipelineError(errorMessage pipeline.PipelineError) error {
	return s.addStep(Step{
		TracingID:      errorMessage.ReceivedByWorker.ID,
		StepIndex:      errorMessage.ReceivedByWorker.StepIndex,
		StepsRemaining: int64(len(errorMessage.ReceivedByWorker.PipelineSteps) - (int(errorMessage.ReceivedByWorker.StepIndex + 1))),
		StartTime:      errorMessage.ReceivedByWorker.Timestamp,
		Error:          errorMessage.Error,
	})
}

func (s *Service) QueryTraces(f Filter, r pagination.Request) (*pagination.Page[TraceDTO], error) {
	page, err := s.stepStore.Query(f, r)
	if err != nil {
		return nil, err
	}

	// All required steps have been retrieved, now each step needs to be ordered by it's corresponding trace ID and send back as a TraceDTO

	return &pagination.Page[TraceDTO]{
		Cursor: page.Cursor,
		Data:   []TraceDTO{},
	}, nil
}

type Filter struct {
	// TODO: some cool stuff...
}

type TraceDTO struct {
	TracingId string    `json:"tracingId"`
	Status    Status    `json:"status"`
	Steps     []StepDTO `json:"steps"`
}

type StepDTO struct {
	Status    Status        `json:"status"`
	StartTime int64         `json:"startTime"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error"`
}

func (s *Service) addStep(step Step) error {
	return s.stepStore.Insert(step)
}
