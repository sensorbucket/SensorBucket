package projects

import (
	"context"
	"fmt"

	"sensorbucket.nl/sensorbucket/pkg/auth"
)

type Store interface {
	CreateProject(context.Context, *Project) error
}

type Application struct {
	store Store
}

func (app *Application) ListProjects(ctx context.Context) {
}

type CreateProjectParams struct {
	Name        string
	Description string
}

func (app *Application) CreateProject(ctx context.Context, params CreateProjectParams) (*Project, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PROJECTS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	project := Project{
		Name:        params.Name,
		Description: params.Description,
		TenantID:    tenantID,
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
	return nil
}

func (app *Application) RemoveProject(ctx context.Context, id int64) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PROJECTS}); err != nil {
		return err
	}
	return nil
}

type AddProjectFeatureOfInterestParams struct {
	FeaturOfInterestID         int64
	InterestedObservationTypes []string
}

func (app *Application) AddProjectFeatureOfInterest(ctx context.Context, params AddProjectFeatureOfInterestParams) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PROJECTS}); err != nil {
		return err
	}
	return nil
}

func (app *Application) RemoveProjectFeatureOfInterest(ctx context.Context) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PROJECTS}); err != nil {
		return err
	}
	return nil
}

type ModifyProjectFeatureOfInterestParams struct {
	FeaturOfInterestID               int64
	AddInterestedObservationTypes    []string
	RemoveInterestedObservationTypes []string
}

func (app *Application) ModifyProjectFeatureOfInterestObservationTypes(ctx context.Context, params ModifyProjectFeatureOfInterestParams) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_PROJECTS}); err != nil {
		return err
	}
	return nil
}
