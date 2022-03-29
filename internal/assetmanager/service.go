package assetmanager

import (
	"fmt"

	"github.com/go-chi/chi/v5"
)

// iService ...
type iService interface {
	ListAssetTypes() ([]string, error)
	RegisterAssetType(AssetSchema) error
	CreateAsset(*Asset) error
	UpdateAsset(*Asset) error
	GetAsset(id string) (*Asset, error)
	DeleteAsset(id string) error
}

var _ iService = (*Service)(nil)

// SchemaRegister ...
type SchemaRegister interface {
	Register(AssetSchema) error
}

// Store ...
type Store interface {
	SchemaRegister
	Create(*Asset) error
	Update(*Asset) error
	Get(urn AssetURN) (*Asset, error)
	Delete(urn AssetURN) error
}

// AssetValidator ...
type AssetValidator interface {
	SchemaRegister
	Validate(*Asset) error
}

// URNGenerator ...
type URNGenerator interface {
	Generate(*Asset) (AssetURN, error)
}

// Service ...
type Service struct {
	store     Store
	validator AssetValidator
	urn       URNGenerator

	assetTypes []string
	router     chi.Router
}

// Opts ...
type Opts struct {
	Store Store
}

func New(opts Opts) *Service {
	svc := &Service{
		store:     opts.Store,
		validator: newSchemaRegistry(),
		urn:       newURNGenerator(),
		router:    chi.NewRouter(),
	}
	svc.setupRoutes()

	return svc
}

func (svc *Service) RegisterAssetType(schema AssetSchema) error {
	if err := svc.validator.Register(schema); err != nil {
		return err
	}
	if err := svc.store.Register(schema); err != nil {
		return err
	}

	svc.assetTypes = append(svc.assetTypes, schema.Type)

	return nil
}

func (svc *Service) ListAssetTypes() ([]string, error) {
	return svc.assetTypes, nil
}

func (svc *Service) CreateAsset(asset *Asset) error {
	if err := svc.validator.Validate(asset); err != nil {
		return fmt.Errorf("could not create asset: %w", err)
	}

	// Generate a corresponding URN
	urn, err := svc.urn.Generate(asset)
	if err != nil {
		return fmt.Errorf("could not create asset: %w", err)
	}
	asset.URN = urn.String()

	if err := svc.store.Create(asset); err != nil {
		return fmt.Errorf("could not create asset: %w", err)
	}
	return nil
}

func (svc *Service) UpdateAsset(asset *Asset) error {
	if err := svc.validator.Validate(asset); err != nil {
		return fmt.Errorf("could not create asset: %w", err)
	}

	if err := svc.store.Update(asset); err != nil {
		return fmt.Errorf("could not create asset: %w", err)
	}
	return nil
}

func (svc *Service) GetAsset(id string) (*Asset, error) {
	urn, err := ParseAssetURN(id)
	if err != nil {
		return nil, fmt.Errorf("could not get asset: %w", err)
	}

	asset, err := svc.store.Get(urn)
	if err != nil {
		return nil, fmt.Errorf("could not get asset: %w", err)
	}
	return asset, nil
}

func (svc *Service) DeleteAsset(id string) error {
	urn, err := ParseAssetURN(id)
	if err != nil {
		return fmt.Errorf("could not delete asset: %w", err)
	}

	if err := svc.store.Delete(urn); err != nil {
		return fmt.Errorf("could not delete asset: %w", err)
	}
	return nil
}
