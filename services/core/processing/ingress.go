package processing

import (
	"time"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type IngressDTO struct {
	TracingID  uuid.UUID `json:"tracing_id"`
	PipelineID uuid.UUID `json:"pipeline_id"`
	OwnerID    int64     `json:"owner_id"`
	Payload    []byte    `json:"payload,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func CreateIngressDTO(pipeline uuid.UUID, ownerID int64, payload []byte) IngressDTO {
	return IngressDTO{
		TracingID:  uuid.New(),
		CreatedAt:  time.Now(),
		PipelineID: pipeline,
		OwnerID:    ownerID,
		Payload:    payload,
	}
}

func TransformIngressDTOToPipelineMessage(dto IngressDTO, pl *Pipeline) (*pipeline.Message, error) {
	pipelineMessage := pipeline.Message{
		ID:            dto.TracingID.String(),
		OwnerID:       dto.OwnerID,
		ReceivedAt:    dto.CreatedAt.UnixMilli(),
		Timestamp:     dto.CreatedAt.UnixMilli(),
		Payload:       dto.Payload,
		PipelineID:    pl.ID,
		PipelineSteps: pl.Steps,
		StepIndex:     0,
		Measurements:  []pipeline.Measurement{},
		Metadata:      make(map[string]any),
	}
	return &pipelineMessage, nil
}
