package ingressarchiver

//go:generate moq -pkg ingressarchiver_test -out mock_test.go . Store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

type Store interface {
	Save(ArchivedIngressDTO) error
	List(ArchiveFilters, pagination.Request) (*pagination.Page[ArchivedIngressDTO], error)
}

type Application struct {
	store Store
}

func New(store Store) *Application {
	return &Application{store}
}

func (a *Application) ArchiveIngressDTO(tracingID uuid.UUID, rawMessage []byte) error {
	archivedDTO := ArchiveIngressDTO(tracingID, rawMessage)
	if err := a.store.Save(archivedDTO); err != nil {
		return fmt.Errorf("error archiving Ingress DTO, store error: %w", err)
	}
	return nil
}

type ArchiveFilters struct {
	TenantID int64
}

func (a *Application) ListIngresses(ctx context.Context, filters ArchiveFilters, p pagination.Request) (*pagination.Page[ArchivedIngressDTO], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	filters.TenantID = tenantID

	return a.store.List(filters, p)
}
