package processing

import (
	"time"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type IngressDTO struct {
	TracingID  uuid.UUID
	PipelineID uuid.UUID
	AuthToken  string
	Payload    []byte
	CreatedAt  time.Time
}

func CreateIngressDTO(pipeline uuid.UUID, auth string, payload []byte) IngressDTO {
	return IngressDTO{
		TracingID:  uuid.New(),
		CreatedAt:  time.Now(),
		PipelineID: pipeline,
		AuthToken:  auth,
		Payload:    payload,
	}
}

func TransformIngressDTOToPipelineMessage(dto IngressDTO, pl *Pipeline) (*pipeline.Message, error) {
	pipelineMessage := pipeline.Message{
		ID:            dto.TracingID.String(),
		ReceivedAt:    dto.CreatedAt.UnixMilli(),
		Timestamp:     dto.CreatedAt.UnixMilli(),
		Payload:       dto.Payload,
		PipelineID:    pl.ID,
		PipelineSteps: pl.Steps,
		Measurements:  []pipeline.Measurement{},
		Metadata:      make(map[string]any),
	}
	return &pipelineMessage, nil
}
