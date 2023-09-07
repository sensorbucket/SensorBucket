package ingressarchiver

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

var (
	_  Store = (*StorePSQL)(nil)
	pq       = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

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

type ArchivedIngressPaginationQuery struct {
	ArchivedAt time.Time `pagination:"archived_at,DESC"`
	TracingID  uuid.UUID `pagination:"tracing_id,ASC"`
}

func (s *StorePSQL) List(filters ArchiveFilters, pageRequest pagination.Request) (*pagination.Page[ArchivedIngressDTO], error) {
	var err error
	q := pq.Select(
		"tracing_id", "raw_message",
		"dto_owner_id", "dto_pipeline_id", "dto_payload", "dto_created_at",
		"archived_at", "expires_at",
	).From("archived_ingress_dtos")

	// Apply pagination
	cursor, err := pagination.GetCursor[ArchivedIngressPaginationQuery](pageRequest)
	if err != nil {
		return nil, fmt.Errorf("list archives, error getting pagination cursor: %w", err)
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, fmt.Errorf("list archives, could not apply pagination: %w", err)
	}

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, fmt.Errorf("list archives, could not run query: %w", err)
	}
	archives := []ArchivedIngressDTO{}
	for rows.Next() {
		var ingress ArchivedIngressDTO
		var dtoOwnerID *int64
		var dtoPipelineID *uuid.UUID
		var dtoPayload []byte
		var dtoCreatedAt *time.Time
		err := rows.Scan(
			&ingress.TracingID, &ingress.RawMessage,
			&dtoOwnerID, &dtoPipelineID, &dtoPayload, &dtoCreatedAt,
			&ingress.ArchivedAt, &ingress.ExpiresAt,
			&cursor.Columns.ArchivedAt, &cursor.Columns.TracingID,
		)
		if err != nil {
			return nil, fmt.Errorf("list archives, could not scan archive: %w", err)
		}
		if dtoOwnerID != nil && dtoPipelineID != nil && dtoPayload != nil && dtoCreatedAt != nil {
			ingress.IngressDTO = &processing.IngressDTO{
				TracingID:  ingress.TracingID,
				OwnerID:    *dtoOwnerID,
				PipelineID: *dtoPipelineID,
				Payload:    dtoPayload,
				CreatedAt:  *dtoCreatedAt,
			}
		}
		archives = append(archives, ingress)
	}

	page := pagination.CreatePageT(archives, cursor)
	return &page, nil
}
