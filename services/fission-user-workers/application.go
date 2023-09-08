package main

import (
	"context"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/pagination"
)

type Store interface {
	WorkersExists([]uuid.UUID) ([]uuid.UUID, error)
	ListUserWorkers(req pagination.Request) (*pagination.Page[UserWorker], error)
	CreateWorker(*UserWorker) error
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
	Source      []byte `json:"source"`
}

func (app *Application) CreateWorker(ctx context.Context, opts CreateWorkerOpts) (*UserWorker, error) {
	worker, err := CreateWorker(opts.Name, opts.Description, opts.Source)
	if err != nil {
		return nil, err
	}
	if err := app.store.CreateWorker(worker); err != nil {
		return nil, err
	}
	return worker, nil
}

func (app *Application) ListWorkers(ctx context.Context, req pagination.Request) (*pagination.Page[UserWorker], error) {
	return app.store.ListUserWorkers(req)
}
