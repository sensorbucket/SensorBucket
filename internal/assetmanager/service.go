package assetmanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	ErrAssetDefinitionNotFound         = errors.New("asset definition not found")
	ErrAssetNotFound                   = errors.New("asset not found")
	ErrExistingAssetDefinitionMismatch = errors.New("provided asset definition schema does not match the existing asset definition in the database")

	ASSET_ID_LENGTH = 8
)

// iService ...
type iService interface {
	ListAssetDefinitions() ([]string, error)
	RegisterAssetDefinition(RegisterAssetDefinitionOpts) error
	GetAssetDefinition(name string) (*AssetDefinition, error)

	CreateAsset(CreateAssetOpts) (*Asset, error)
	UpdateAsset(UpdateAssetOpts) error
	GetAsset(urnString string) (*Asset, error)
	DeleteAsset(urnString string) error
	FindAssets(definitionURN string, filter map[string]interface{}) ([]Asset, error)
}

var _ iService = (*Service)(nil)

// AssetDefinitionStore ...
type AssetDefinitionStore interface {
	CreateAssetDefinition(*AssetDefinition) error
	GetAssetDefinition(urn AssetURN) (*AssetDefinition, error)
}

// AssetStore ...
type AssetStore interface {
	CreateAsset(*Asset) error
	UpdateAsset(*Asset) error
	GetAsset(urn AssetURN) (*Asset, error)
	DeleteAsset(urn AssetURN) error
	FindFilter(definitionURN AssetURN, filter map[string]interface{}) ([]Asset, error)
}

// Service ...
type Service struct {
	atStore    AssetDefinitionStore
	aStore     AssetStore
	pipelineID string

	assetDefinitions []string
	router           chi.Router
}

// Opts ...
type Opts struct {
	AssetStore           AssetStore
	AssetDefinitionStore AssetDefinitionStore
	PipelineID           string
}

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

// RegisterAssetDefinitionOpts ...
type RegisterAssetDefinitionOpts struct {
	Name       string
	PipelineID string // Shouldn't the service specify the PipelineID
	Labels     []string
	Version    int
	Schema     json.RawMessage
}

func (svc *Service) RegisterAssetDefinition(opts RegisterAssetDefinitionOpts) error {
	at, err := newAssetDefinition(newAssetDefinitionOpts{
		Name:       opts.Name,
		PipelineID: opts.PipelineID,
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

	if err := svc.aStore.CreateAsset(asset); err != nil {
		return nil, fmt.Errorf("could not create asset: %w", err)
	}
	return asset, nil
}

// UpdateAssetOpts ...
type UpdateAssetOpts struct {
}

func (svc *Service) UpdateAsset(opts UpdateAssetOpts) error {
	return fmt.Errorf("not implemented")
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

// var randomChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
// func randomString(length int) string {
// 	b := make([]byte, length)
// 	rand.Read(b)
// 	for i, v := range b {
// 		b[i] = randomChars[v%byte(len(randomChars))]
// 	}
// 	return string(b)
// }

func randomString(length int) string {
	return gonanoid.Must(length)
}

/*

64 ^ 8



*/
