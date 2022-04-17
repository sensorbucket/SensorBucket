package assetmanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	ASSET_ID_LENGTH = 8
)

var (
	ErrAssetDefinitionNotFound         = errors.New("asset definition not found")
	ErrAssetNotFound                   = errors.New("asset not found")
	ErrDuplicateAssetID                = errors.New("duplicate asset id")
	ErrExistingAssetDefinitionMismatch = errors.New("provided asset definition schema does not match the existing asset definition in the database")
)

// iService is the interface for this service. It is not used except for the developer to get a quick overview of existing functions
type iService interface {
	ListAssetDefinitions() ([]string, error)
	RegisterAssetDefinition(RegisterAssetDefinitionOpts) error
	GetAssetDefinition(name string) (*AssetDefinition, error)

	CreateAsset(CreateAssetOpts) (*Asset, error)
	UpdateAsset(urn string, content json.RawMessage) error
	GetAsset(urnString string) (*Asset, error)
	DeleteAsset(urnString string) error
	FindAssets(definitionURN string, filter map[string]interface{}) ([]Asset, error)
}

// Validate that Service implements this interface
var _ iService = (*Service)(nil)

// AssetDefinitionStore stores asset definitions
type AssetDefinitionStore interface {
	CreateAssetDefinition(*AssetDefinition) error
	GetAssetDefinition(urn AssetURN) (*AssetDefinition, error)
}

// AssetStore stores assets
type AssetStore interface {
	CreateAsset(*Asset) error
	UpdateAsset(*Asset) error
	GetAsset(urn AssetURN) (*Asset, error)
	DeleteAsset(urn AssetURN) error
	FindFilter(definitionURN AssetURN, filter map[string]interface{}) ([]Asset, error)
}

// Service is the AssetManager service that allows custom assets to be defined and allows CRUD operations on them
type Service struct {
	atStore    AssetDefinitionStore
	aStore     AssetStore
	pipelineID string

	assetDefinitions []string
	router           chi.Router
}

// Opts Options required during creation of the service
type Opts struct {
	AssetStore           AssetStore
	AssetDefinitionStore AssetDefinitionStore
	PipelineID           string
}

// New creates a new asset manager service
func New(opts Opts) *Service {
	svc := &Service{
		aStore:     opts.AssetStore,
		atStore:    opts.AssetDefinitionStore,
		pipelineID: opts.PipelineID,
		router:     chi.NewRouter(),
	}
	svc.setupRoutes()

	return svc
}

// RegisterAssetDefinitionOpts Options required to register a new asset definition
type RegisterAssetDefinitionOpts struct {
	Name       string
	PipelineID string // Shouldn't the service specify the PipelineID
	PrimaryKey string // A path to the key in the asset that will be used in the URN, defaults to random string
	Labels     []string
	Version    int
	Schema     json.RawMessage
}

// RegisterAssetDefinition registers a new asset definition
func (svc *Service) RegisterAssetDefinition(opts RegisterAssetDefinitionOpts) error {
	at, err := newAssetDefinition(newAssetDefinitionOpts{
		Name:       opts.Name,
		PipelineID: opts.PipelineID,
		PrimaryKey: opts.PrimaryKey,
		Labels:     []string{},
		Version:    opts.Version,
		Schema:     opts.Schema,
	})

	if err != nil {
		return fmt.Errorf("could not create asset definition: %w", err)
	}

	// Assert the asset definition
	existingAT, err := svc.atStore.GetAssetDefinition(at.URN())
	if err != nil && !errors.Is(err, ErrAssetDefinitionNotFound) {
		return err
	}
	if errors.Is(err, ErrAssetDefinitionNotFound) {
		return svc.atStore.CreateAssetDefinition(at)
	}
	if !existingAT.Equals(at) {
		return ErrExistingAssetDefinitionMismatch
	}

	svc.assetDefinitions = append(svc.assetDefinitions, at.URN().String())

	return nil
}

func (svc *Service) GetAssetDefinition(urnString string) (*AssetDefinition, error) {
	urn, err := ParseAssetURN(urnString)
	if err != nil {
		return nil, err
	}

	return svc.atStore.GetAssetDefinition(urn)
}

func (svc *Service) ListAssetDefinitions() ([]string, error) {
	return svc.assetDefinitions, nil
}

type CreateAssetOpts struct {
	AssetDefinition string
	Content         json.RawMessage
}

func (svc *Service) CreateAsset(opts CreateAssetOpts) (*Asset, error) {
	at, err := svc.GetAssetDefinition(opts.AssetDefinition)
	if err != nil {
		return nil, err
	}

	asset, err := newAsset(at, opts.Content)
	if err != nil {
		return nil, err
	}

	if err := asset.Validate(); err != nil {
		return nil, fmt.Errorf("could not validate asset: %w", err)
	}

	if exists, err := svc.assetExists(asset.URN()); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrDuplicateAssetID
	}

	if err := svc.aStore.CreateAsset(asset); err != nil {
		return nil, fmt.Errorf("could not create asset: %w", err)
	}
	return asset, nil
}

func (svc *Service) assetExists(urn AssetURN) (bool, error) {
	_, err := svc.aStore.GetAsset(urn)
	if err != nil {
		if errors.Is(err, ErrAssetNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UpdateAssetOpts options required for updating an asset
type UpdateAssetOpts struct {
	URN     string
	Content json.RawMessage
}

func (svc *Service) UpdateAsset(urn string, content json.RawMessage) error {
	asset, err := svc.GetAsset(urn)
	if err != nil {
		return err
	}

	asset.Content = content

	if err := asset.Validate(); err != nil {
		return fmt.Errorf("could not validate asset: %w", err)
	}

	if err := svc.aStore.UpdateAsset(asset); err != nil {
		return fmt.Errorf("could not update asset: %w", err)
	}

	return nil
}

func (svc *Service) GetAsset(urnString string) (*Asset, error) {
	urn, err := ParseAssetURN(urnString)
	if err != nil {
		return nil, err
	}

	asset, err := svc.aStore.GetAsset(urn)
	if err != nil {
		return nil, fmt.Errorf("could not get asset: %w", err)
	}
	return asset, nil
}

func (svc *Service) DeleteAsset(urnString string) error {
	urn, err := ParseAssetURN(urnString)
	if err != nil {
		return err
	}

	if err := svc.aStore.DeleteAsset(urn); err != nil {
		return fmt.Errorf("could not delete asset: %w", err)
	}
	return nil
}

// FindAssets uses a filter to find specific assets where the content matches the filter
// The filter is a key value map, where the key is a top level key in the asset's content
func (svc *Service) FindAssets(definitionURN string, filter map[string]interface{}) ([]Asset, error) {
	urn, err := ParseAssetURN(definitionURN)
	if err != nil {
		return nil, err
	}

	return svc.aStore.FindFilter(urn, filter)
}

func randomString(length int) string {
	return gonanoid.Must(length)
}
