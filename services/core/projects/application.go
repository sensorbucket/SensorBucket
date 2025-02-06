package projects

import (
	"context"
	"fmt"
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

var ErrProjectNotFound = web.NewError(http.StatusNotFound, "project was not found", "PROJECT_NOT_FOUND")

type Store interface {
	ListProjects(context.Context, ProjectsFilter, pagination.Request) (*pagination.Page[*Project], error)
	CreateProject(context.Context, *Project) error
	EditProject(context.Context, EditProjectParams) error
	RemoveProject(context.Context, int64) error
	SetProjectFeatureOfInterest(context.Context, ModifyProjectFeatureOfInterestParams) error
}

type Application struct {
	store Store
}

func New(store Store) *Application {
	return &Application{store: store}
}

type ProjectsFilter struct{}

func (app *Application) ListProjects(ctx context.Context, filter ProjectsFilter, p pagination.Request) (*pagination.Page[*Project], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_PROJECTS}); err != nil {
		return nil, err
	}
	page, err := app.store.ListProjects(ctx, filter, p)
	if err != nil {
		return nil, fmt.Errorf("", err)
	}

	return page, nil
}

type CreateProjectParams struct {
	Name        string
	Description string
}

func (app *Application) CreateProject(ctx context.Context, params CreateProjectParams) (*Project, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PROJECTS}); err != nil {
		return nil, err
	}

	project := Project{
		Name:        params.Name,
		Description: params.Description,
	}

	if err := app.store.CreateProject(ctx, &project); err != nil {
		return nil, fmt.Errorf("could not create project in store: %w", err)
	}
	return &project, nil
}

type EditProjectParams struct {
	ID          int64
	Name        *string
	Description *string
}

func (app *Application) EditProject(ctx context.Context, params EditProjectParams) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PROJECTS}); err != nil {
		return err
	}
	if err := app.store.EditProject(ctx, params); err != nil {
		return fmt.Errorf("", err)
	}
	return nil
}

func (app *Application) RemoveProject(ctx context.Context, id int64) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PROJECTS}); err != nil {
		return err
	}
	if err := app.store.RemoveProject(ctx, id); err != nil {
		return fmt.Errorf("", err)
	}
	return nil
}

type ModifyProjectFeatureOfInterestParams struct {
	FeaturOfInterestID         int64
	InterestedObservationTypes []string
}

func (app *Application) ModifyProjectFeatureOfInterest(ctx context.Context, params ModifyProjectFeatureOfInterestParams) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PROJECTS}); err != nil {
		return err
	}
	// Check if user has access to all the requested feature of interests
	return nil
}
