package processing

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/pkg/web"
)

var (
	ErrPipelineNotFound      = web.NewError(http.StatusNotFound, "pipeline was not found", "PIPELINE_NOT_FOUND")
	ErrPipelineNotActive     = web.NewError(http.StatusNotFound, "pipeline is currently not active", "PIPELINE_NOT_ACTIVE")
	ErrPipelineNotInactive   = web.NewError(http.StatusBadRequest, "pipeline is currently not disabled", "PIPELINE_NOT_DISABLED")
	ErrPipelineInvalidStep   = web.NewError(http.StatusBadRequest, "pipeline step is invalid", "PIPELINE_INVALID_STEP")
	ErrPipelineInvalidStatus = web.NewError(http.StatusBadRequest, "pipeline status is invalid", "PIPELINE_INVALID_STATUS")
)

type PipelineStatus string

const (
	PipelineActive   PipelineStatus = "active"
	PipelineInactive PipelineStatus = "inactive"
)

func StrToStatus(str string) (PipelineStatus, error) {
	switch str {
	case string(PipelineActive):
		return PipelineActive, nil
	case string(PipelineInactive):
		return PipelineInactive, nil
	default:
		return PipelineActive, ErrPipelineInvalidStatus
	}
}

type Pipeline struct {
	ID               string         `json:"id"`
	Description      string         `json:"description"`
	Status           PipelineStatus `json:"status"`
	Steps            []string       `json:"steps"`
	LastStatusChange time.Time      `json:"last_status_change"`
	CreatedAt        time.Time      `json:"created_at"`
}

func NewPipeline(description string, steps []string) (*Pipeline, error) {
	p := &Pipeline{
		ID:          uuid.Must(uuid.NewRandom()).String(),
		Description: description,
		Status:      PipelineActive,
		CreatedAt:   time.Now(),
	}

	if err := p.SetSteps(steps); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Pipeline) SetStatus(status PipelineStatus) error {
	var err error
	switch status {
	case PipelineActive:
		err = p.Enable()
	case PipelineInactive:
		err = p.Disable()
	default:
		err = ErrPipelineInvalidStatus
	}
	return err
}

func (p *Pipeline) Disable() error {
	if p.Status != PipelineActive {
		return fmt.Errorf("cannot disable pipeline: %w", ErrPipelineNotActive)
	}
	p.Status = PipelineInactive
	p.LastStatusChange = time.Now()
	return nil
}

func (p *Pipeline) Enable() error {
	if p.Status != PipelineInactive {
		return fmt.Errorf("cannot enable pipeline: %w", ErrPipelineNotInactive)
	}
	p.Status = PipelineActive
	p.LastStatusChange = time.Now()
	return nil
}

const _R_STEP = "^[a-zA-Z0-9_-]+$"

var R_STEP = regexp.MustCompile(_R_STEP)

func (p *Pipeline) SetSteps(steps []string) error {
	for _, step := range steps {
		if !R_STEP.MatchString(step) {
			return fmt.Errorf("%w: this step: '%s'", ErrPipelineInvalidStep, step)
		}
	}
	p.Steps = steps

	return nil
}
