package processing

import (
	"context"
	"fmt"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/ingress"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type Store interface {
	CreatePipeline(*Pipeline) error
	UpdatePipeline(*Pipeline) error
	ListPipelines(PipelinesFilter, pagination.Request) (pagination.Page[Pipeline], error)
	GetPipeline(string) (*Pipeline, error)
}

type Service struct {
	store                    Store
	pipelineMessageProcessor PipelineMessageProcessor
	measurementStorer        MeasurementStorer
	measurementNotifier      NewMeasurementNotifier
}

func New(store Store, messageProcessor PipelineMessageProcessor) *Service {
	s := &Service{
		store:                    store,
		pipelineMessageProcessor: messageProcessor,
	}
	return s
}

type CreatePipelineDTO struct {
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
}

func (s *Service) CreatePipeline(ctx context.Context, dto CreatePipelineDTO) (*Pipeline, error) {
	p, err := NewPipeline(dto.Description, dto.Steps)
	if err != nil {
		return nil, err
	}
	if err := s.store.CreatePipeline(p); err != nil {
		return nil, err
	}
	return p, nil
}

type PipelinesFilter struct {
	Status []PipelineStatus
	Step   []string
}

func NewPipelinesFilter() PipelinesFilter {
	return PipelinesFilter{}
}

func (s *Service) ListPipelines(ctx context.Context, filter PipelinesFilter, p pagination.Request) (pagination.Page[Pipeline], error) {
	pipelines, err := s.store.ListPipelines(filter, p)
	return pipelines, err
}

func (s *Service) GetPipeline(ctx context.Context, id string, allowInactive bool) (*Pipeline, error) {
	p, err := s.store.GetPipeline(id)
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

// ProcessIngressDTO will intiate the processing of the given ingress data
func (s *Service) ProcessIngressDTO(ctx context.Context, dto ingress.DTO) error {
	pl, err := s.GetPipeline(ctx, dto.PipelineID, false)
	if err != nil {
		return fmt.Errorf("ProcessIngressData could not get pipeline: %w", err)
	}

	// Convert ingress DTO to pipeline message
	return pl.Process(dto.Payload, s.pipelineMessageProcessor)
}

func (s *Service) ProcessPipelineResult(ctx context.Context, msg pipeline.Message) error {
	return nil
}
