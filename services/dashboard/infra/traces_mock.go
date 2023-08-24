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
		steps[ix] = routes.StepDTO{
			Status:   rand.Intn(5),
			Duration: time.Duration(rand.Float32()) * time.Second,
			Error:    "",
		}
	}
	return steps
}

func (t *TracesMock) ListTraces(ids []uuid.UUID) ([]routes.TraceDTO, error) {
	traces := []routes.TraceDTO{}
	for ix, id := range ids {
		traces[ix] = routes.TraceDTO{
			TracingId: id.String(),
			Status:    rand.Intn(5),
			Steps:     genSteps(),
		}
	}
	return traces, nil
}
