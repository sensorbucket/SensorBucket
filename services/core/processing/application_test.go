package processing_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

func TestShouldProcessIngressDTO(t *testing.T) {
	// Arrange
	plID := uuid.New()
	pl := processing.Pipeline{
		ID:    plID.String(),
		Steps: []string{"a", "b", "c"},
	}
	store := &StoreMock{
		GetPipelineFunc: func(s string) (*processing.Pipeline, error) {
			return &pl, nil
		},
	}
	publ := make(chan *pipeline.Message, 10)
	svc := processing.New(store, publ)
	dto := processing.IngressDTO{
		TracingID:  uuid.New(),
		PipelineID: plID,
		CreatedAt:  time.Now(),
		AuthToken:  "123123123",
		Payload:    []byte("Hello world"),
	}

	// Act
	err := svc.ProcessIngressDTO(context.Background(), dto)

	// Assert
	assert.NoError(t, err)
	require.Len(t, publ, 1)
	result := <-publ
	assert.Equal(t, dto.TracingID.String(), result.ID)
	assert.Equal(t, pl.Steps, result.PipelineSteps)
	assert.Equal(t, dto.Payload, result.Payload)
}
