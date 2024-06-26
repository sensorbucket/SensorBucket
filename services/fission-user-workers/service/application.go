package userworkers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

var (
	ErrWorkerNotFound     = web.NewError(http.StatusNotFound, "Worker was not found", "ERR_WORKER_NOT_FOUND")
	ErrInvalidUUID        = web.NewError(http.StatusBadRequest, "given UUID was invalid", "ERR_UUID_INVALID")
	ErrWorkerInvalidState = web.NewError(http.StatusBadRequest, "given worker state was invalid", "ERR_WORKER_STATE_INVALID")
)

type Store interface {
	WorkersExists([]uuid.UUID, WorkerFilters) ([]uuid.UUID, error)
	ListUserWorkers(WorkerFilters, pagination.Request) (*pagination.Page[UserWorker], error)
	CreateWorker(*UserWorker) error
	GetWorkerByID(uuid.UUID, WorkerFilters) (*UserWorker, error)
	UpdateWorker(*UserWorker) error
}

type Application struct {
	store Store
}

func NewApplication(store Store) *Application {
	return &Application{store}
}

type CreateWorkerOpts struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	State       WorkerState `json:"state"`
	UserCode    []byte      `json:"user_code"`
}

func (app *Application) GetWorker(ctx context.Context, id uuid.UUID) (*UserWorker, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_USER_WORKERS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return app.store.GetWorkerByID(id, WorkerFilters{TenantID: []int64{tenantID}})
}

func (app *Application) CreateWorker(ctx context.Context, opts CreateWorkerOpts) (*UserWorker, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_USER_WORKERS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	worker, err := CreateWorker(tenantID, opts.Name, opts.Description, opts.UserCode)
	if err != nil {
		return nil, err
	}
	if opts.State == StateEnabled {
		worker.Enable()
	}
	if err := app.store.CreateWorker(worker); err != nil {
		return nil, err
	}
	return worker, nil
}

type UpdateWorkerOpts struct {
	Name        *string      `json:"name"`
	Description *string      `json:"description"`
	UserCode    []byte       `json:"user_code"`
	State       *WorkerState `json:"state"`
}

func (app *Application) UpdateWorker(ctx context.Context, worker *UserWorker, opts UpdateWorkerOpts) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_USER_WORKERS}); err != nil {
		return err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return err
	}
	if worker.TenantID != tenantID {
		return auth.ErrUnauthorized
	}

	if opts.Name != nil {
		if err := worker.SetName(*opts.Name); err != nil {
			return err
		}
	}
	if opts.Description != nil {
		worker.Description = strings.Trim(*opts.Description, "\r\n\t ")
	}
	if opts.State != nil {
		switch *opts.State {
		case StateEnabled:
			worker.Enable()
		case StateDisabled:
			worker.Disable()
		default:
			worker.Disable()
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

type WorkerFilters struct {
	ID       []string
	State    WorkerState
	TenantID []int64
}

func (app *Application) ListWorkers(ctx context.Context, filters WorkerFilters, req pagination.Request) (*pagination.Page[UserWorker], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_USER_WORKERS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	filters.TenantID = []int64{tenantID}

	return app.store.ListUserWorkers(filters, req)
}
