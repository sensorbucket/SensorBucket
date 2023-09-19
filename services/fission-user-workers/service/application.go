package userworkers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrWorkerNotFound     = web.NewError(http.StatusNotFound, "Worker was not found", "ERR_WORKER_NOT_FOUND")
	ErrInvalidUUID        = web.NewError(http.StatusBadRequest, "given UUID was invalid", "ERR_UUID_INVALID")
	ErrWorkerInvalidState = web.NewError(http.StatusBadRequest, "given worker state was invalid", "ERR_WORKER_STATE_INVALID")
)

type Store interface {
	WorkersExists([]uuid.UUID, ListWorkerFilters) ([]uuid.UUID, error)
	ListUserWorkers(ListWorkerFilters, pagination.Request) (*pagination.Page[UserWorker], error)
	CreateWorker(*UserWorker) error
	GetWorkerByID(uuid.UUID) (*UserWorker, error)
	UpdateWorker(*UserWorker) error
}

type Application struct {
	store Store
}

func NewApplication(store Store) *Application {
	return &Application{store}
}

type CreateWorkerOpts struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserCode    []byte `json:"user_code"`
}

func (app *Application) GetWorker(ctx context.Context, id uuid.UUID) (*UserWorker, error) {
	return app.store.GetWorkerByID(id)
}

func (app *Application) CreateWorker(ctx context.Context, opts CreateWorkerOpts) (*UserWorker, error) {
	worker, err := CreateWorker(opts.Name, opts.Description, opts.UserCode)
	if err != nil {
		return nil, err
	}
	if err := app.store.CreateWorker(worker); err != nil {
		return nil, err
	}
	return worker, nil
}

type UpdateWorkerOpts struct {
	UserCode    []byte       `json:"user_code"`
	Description *string      `json:"description"`
	State       *WorkerState `json:"state"`
}

func (app *Application) UpdateWorker(ctx context.Context, worker *UserWorker, opts UpdateWorkerOpts) error {
	if opts.Description != nil {
		worker.Description = *opts.Description
	}
	if opts.State != nil {
		switch *opts.State {
		case StateEnabled:
			worker.Enable()
		case StateDisabled:
			worker.Disable()
		default:
			return ErrWorkerInvalidState
		}
	}
	if opts.UserCode != nil {
		err := worker.SetUserCode(opts.UserCode)
		if err != nil {
			return fmt.Errorf("error setting worker source code: %w", err)
		}
	}
	worker.Commit()
	if err := app.store.UpdateWorker(worker); err != nil {
		return fmt.Errorf("could not update worker in db: %w", err)
	}

	return nil
}

type ListWorkerFilters struct {
	ID    []string
	State WorkerState
}

func (app *Application) ListWorkers(ctx context.Context, filters ListWorkerFilters, req pagination.Request) (*pagination.Page[UserWorker], error) {
	return app.store.ListUserWorkers(filters, req)
}
