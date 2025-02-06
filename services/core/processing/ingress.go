package processing

import (
	"time"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type IngressDTO struct {
	TracingID   uuid.UUID      `json:"tracing_id"`
	PipelineID  uuid.UUID      `json:"pipeline_id"`
	TenantID    int64          `json:"tenant_id"`
	Payload     []byte         `json:"payload,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	AccessToken string         `json:"access_token"`
	Metadata    map[string]any `json:"metadata"`
}

func CreateIngressDTO(accessToken string, pipeline uuid.UUID, ownerID int64, payload []byte) IngressDTO {
	return IngressDTO{
		TracingID:   uuid.Must(uuid.NewV7()),
		CreatedAt:   time.Now(),
		PipelineID:  pipeline,
		TenantID:    ownerID,
		Payload:     payload,
		AccessToken: accessToken,
		Metadata:    make(map[string]any),
	}
}

func TransformIngressDTOToPipelineMessage(dto IngressDTO, pl *Pipeline) (*pipeline.Message, error) {
	pipelineMessage := pipeline.Message{
		TracingID:     dto.TracingID.String(),
		AccessToken:   dto.AccessToken,
		TenantID:      dto.TenantID,
		ReceivedAt:    dto.CreatedAt.UnixMilli(),
		Timestamp:     dto.CreatedAt.UnixMilli(),
		Payload:       dto.Payload,
		PipelineID:    pl.ID,
		PipelineSteps: pl.Steps,
		StepIndex:     0,
		Measurements:  []pipeline.Measurement{},
		Metadata:      dto.Metadata,
	}
	return &pipelineMessage, nil
}
