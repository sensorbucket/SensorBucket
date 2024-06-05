package processing

//go:generate moq -pkg processing_test -out mock_test.go . Store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type Store interface {
	CreatePipeline(*Pipeline) error
	UpdatePipeline(*Pipeline) error
	ListPipelines(PipelinesFilter, pagination.Request) (pagination.Page[Pipeline], error)
	GetPipeline(string, PipelinesFilter) (*Pipeline, error)
}

type Service struct {
	store                    Store
	pipelineMessagePublisher PipelineMessagePublisher
	keyClient                auth.JWKSClient
}

func New(store Store, publisher PipelineMessagePublisher, keyClient auth.JWKSClient) *Service {
	s := &Service{
		store:                    store,
		pipelineMessagePublisher: publisher,
		keyClient:                keyClient,
	}
	return s
}

type CreatePipelineDTO struct {
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
}

func (s *Service) CreatePipeline(ctx context.Context, dto CreatePipelineDTO) (*Pipeline, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PIPELINES}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	p, err := NewPipeline(tenantID, dto.Description, dto.Steps)
	if err != nil {
		return nil, err
	}
	if err := s.store.CreatePipeline(p); err != nil {
		return nil, err
	}
	return p, nil
}

type PipelinesFilter struct {
	ID       []uuid.UUID
	TenantID []int64
	Status   []PipelineStatus
	Step     []string
}

func NewPipelinesFilter() PipelinesFilter {
	return PipelinesFilter{}
}

func (s *Service) ListPipelines(ctx context.Context, filter PipelinesFilter, p pagination.Request) (*pagination.Page[Pipeline], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_PIPELINES}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	filter.TenantID = []int64{tenantID}

	pipelines, err := s.store.ListPipelines(filter, p)
	return &pipelines, err
}

// TODO: id should be a UUID!
func (s *Service) GetPipeline(ctx context.Context, id string, allowInactive bool) (*Pipeline, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_PIPELINES}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)

	p, err := s.store.GetPipeline(id, PipelinesFilter{TenantID: []int64{tenantID}})
	if err != nil {
		return nil, err
	}

	if !allowInactive && p.Status == PipelineInactive {
		return nil, ErrPipelineNotActive
	}

	return p, nil
}

type UpdatePipelineDTO struct {
	Description *string         `json:"description,omitempty"`
	Steps       []string        `json:"steps,omitempty"`
	Status      *PipelineStatus `json:"status,omitempty"`
}

func (s *Service) UpdatePipeline(ctx context.Context, id string, dto UpdatePipelineDTO) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PIPELINES}); err != nil {
		return err
	}

	p, err := s.GetPipeline(ctx, id, true)
	if err != nil {
		return err
	}

	if dto.Description != nil {
		p.Description = *dto.Description
	}
	if dto.Steps != nil {
		if err := p.SetSteps(dto.Steps); err != nil {
			return err
		}
	}
	if dto.Status != nil {
		if err := p.SetStatus(*dto.Status); err != nil {
			return err
		}
	}

	if err := s.store.UpdatePipeline(p); err != nil {
		return err
	}
	return nil
}

func (s *Service) DisablePipeline(ctx context.Context, id string) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PIPELINES}); err != nil {
		return err
	}

	p, err := s.GetPipeline(ctx, id, false)
	if err != nil {
		return err
	}
	if err := p.Disable(); err != nil {
		return err
	}
	if err := s.store.UpdatePipeline(p); err != nil {
		return err
	}
	return nil
}

func (s *Service) EnablePipeline(ctx context.Context, id string) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PIPELINES}); err != nil {
		return err
	}

	p, err := s.GetPipeline(ctx, id, true)
	if err != nil {
		return err
	}
	if err := p.Enable(); err != nil {
		return err
	}
	if err := s.store.UpdatePipeline(p); err != nil {
		return err
	}
	return nil
}

type PipelineMessagePublisher chan<- *pipeline.Message

func (s *Service) ProcessIngressDTO(dto IngressDTO) error {
	ctx, err := auth.AuthenticateContext(context.Background(), dto.AccessToken, s.keyClient)
	if err != nil {
		return err
	}
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_MEASUREMENTS}); err != nil {
		return err
	}

	pl, err := s.GetPipeline(ctx, dto.PipelineID.String(), false)
	if err != nil {
		return fmt.Errorf("cannot get pipeline for dto: %w", err)
	}

	pipelineMessage, err := TransformIngressDTOToPipelineMessage(dto, pl)
	if err != nil {
		return fmt.Errorf("cannot transform dto to pipeline Message: %w", err)
	}

	s.pipelineMessagePublisher <- pipelineMessage

	return nil
}
