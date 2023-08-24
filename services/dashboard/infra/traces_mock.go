package dashboardinfra

import (
	"math/rand"
	"time"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/services/dashboard/routes"
)

type TracesMock struct{}

func NewTracesMock() *TracesMock {
	return &TracesMock{}
}

func genSteps() []routes.StepDTO {
	steps := make([]routes.StepDTO, 3)
	for ix := range steps {
		step := routes.StepDTO{
			Status: rand.Intn(5),
		}
		if step.Status == 4 {
			step.Error = "Random error"
		} else {
			step.Duration = time.Duration(rand.Float32() * float32(time.Second))
		}
		steps[ix] = step
	}
	return steps
}

func (t *TracesMock) ListTraces(ids []uuid.UUID) ([]routes.TraceDTO, error) {
	traces := make([]routes.TraceDTO, len(ids))
	for ix, id := range ids {
		traces[ix] = routes.TraceDTO{
			TracingId: id.String(),
			Status:    rand.Intn(5),
			Steps:     genSteps(),
			DeviceID:  int64(rand.Intn(60)),
		}
	}
	return traces, nil
}
