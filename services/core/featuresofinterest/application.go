package featuresofinterest

import (
	"context"
	"encoding/json"
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

var ErrFeatureOfInterestNotFound = web.NewError(http.StatusNotFound, "The requested feature of interest was not found", "FEATURE_OF_INTEREST_NOT_FOUND")

type Service struct {
	store Store
}
type Store interface {
	ListFeaturesOfInterest(ctx context.Context, filter FeatureOfInterestFilter, pageReq pagination.Request) (*pagination.Page[FeatureOfInterest], error)
	GetFeatureOfInterest(ctx context.Context, id int64, filter FeatureOfInterestFilter) (*FeatureOfInterest, error)
	DeleteFeatureOfInterest(ctx context.Context, id int64) error
	SaveFeatureOfInterest(ctx context.Context, foi *FeatureOfInterest) error
}

func NewService(store Store) *Service {
	return &Service{store}
}

type FeatureOfInterestFilter struct {
	TenantID []int64
}

func (service *Service) ListFeaturesOfInterest(
	ctx context.Context, filter FeatureOfInterestFilter, paginationParameters pagination.Request,
) (*pagination.Page[FeatureOfInterest], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	filter.TenantID = []int64{tenantID}
	return service.store.ListFeaturesOfInterest(ctx, filter, paginationParameters)
}

func (service *Service) GetFeatureOfInterest(ctx context.Context, id int64) (*FeatureOfInterest, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	return service.store.GetFeatureOfInterest(ctx, id, FeatureOfInterestFilter{TenantID: []int64{tenantID}})
}

func (service *Service) CreateFeatureOfInterest(ctx context.Context, opts CreateFeatureOfInterestOpts) (*FeatureOfInterest, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	opts.TenantID = tenantID
	foi, err := NewFeatureOfInterest(opts)
	if err != nil {
		return nil, err
	}
	if err := service.store.SaveFeatureOfInterest(ctx, foi); err != nil {
		return nil, err
	}
	return foi, nil
}

func (service *Service) DeleteFeatureOfInterest(ctx context.Context, id int64) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return err
	}
	if _, err := service.GetFeatureOfInterest(ctx, id); err != nil {
		return err
	}
	return service.store.DeleteFeatureOfInterest(ctx, id)
}

type UpdateFeatureOfInterestOpts struct {
	Name         *string
	Description  *string
	EncodingType *string
	Feature      *Geometry
	Properties   json.RawMessage
}

func (service *Service) UpdateFeatureOfInterest(ctx context.Context, id int64, opts UpdateFeatureOfInterestOpts) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return err
	}

	foi, err := service.GetFeatureOfInterest(ctx, id)
	if err != nil {
		return err
	}

	if opts.Name != nil {
		foi.Name = *opts.Name
	}
	if opts.Description != nil {
		foi.Description = *opts.Description
	}
	if opts.Feature != nil {
		foi.Feature = opts.Feature
	}
	if opts.Properties != nil {
		foi.Properties = opts.Properties
	}

	return service.store.SaveFeatureOfInterest(ctx, foi)
}
