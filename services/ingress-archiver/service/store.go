package ingressarchiver

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var _ Store = (*StorePSQL)(nil)

type StorePSQL struct {
	db *sqlx.DB
}

func NewStorePSQL(db *sqlx.DB) *StorePSQL {
	return &StorePSQL{db}
}

func (s *StorePSQL) Save(dto ArchivedIngressDTO) error {
	var dtoOwnerID *int64
	var dtoPipeline *uuid.UUID
	var dtoPayload []byte
	var dtoCreatedAt *time.Time

	if dto.IngressDTO != nil {
		dtoOwnerID = &dto.IngressDTO.OwnerID
		dtoPipeline = &dto.IngressDTO.PipelineID
		dtoPayload = dto.IngressDTO.Payload
		dtoCreatedAt = &dto.IngressDTO.CreatedAt
	}

	_, err := s.db.Exec(`
        INSERT INTO archived_ingress_dtos (
            tracing_id, raw_message,
            dto_owner_id, dto_pipeline_id, dto_payload, dto_created_at,
            archived_at, expires_at
        )
        VALUES (
            $1, $2,
            $3, $4, $5, $6,
            $7, $8
        );
        `,
		dto.TracingID, dto.RawMessage,
		dtoOwnerID, dtoPipeline, dtoPayload, dtoCreatedAt,
		dto.ArchivedAt, dto.ExpiresAt)
	return err
}
