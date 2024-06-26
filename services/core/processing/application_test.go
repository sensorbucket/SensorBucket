package processing_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/pkg/authtest"
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
		GetPipelineFunc: func(s string, filter processing.PipelinesFilter) (*processing.Pipeline, error) {
			return &pl, nil
		},
	}
	publ := make(chan *pipeline.Message, 10)
	svc := processing.New(store, publ, authtest.JWKS())
	dto := processing.IngressDTO{
		TracingID:   uuid.New(),
		AccessToken: authtest.CreateToken(),
		PipelineID:  plID,
		CreatedAt:   time.Now(),
		TenantID:    123,
		Payload:     []byte("Hello world"),
	}

	// Act
	err := svc.ProcessIngressDTO(dto)

	// Assert
	assert.NoError(t, err)
	require.Len(t, publ, 1)
	result := <-publ
	assert.Equal(t, dto.TracingID.String(), result.TracingID)
	assert.Equal(t, pl.Steps, result.PipelineSteps)
	assert.Equal(t, dto.Payload, result.Payload)
	assert.Equal(t, dto.TenantID, result.TenantID)
}
