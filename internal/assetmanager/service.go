package assetmanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	ErrAssetTypeNotFound         = errors.New("asset type not found")
	ErrAssetNotFound             = errors.New("asset not found")
	ErrExistingAssetTypeMismatch = errors.New("provided asset type schema does not match the existing asset type in the database")

	ASSET_ID_LENGTH = 8
)

// iService ...
type iService interface {
	ListAssetTypes() ([]string, error)
	RegisterAssetType(RegisterAssetTypeOpts) error
	GetAssetType(name string) (*AssetType, error)

	CreateAsset(CreateAssetOpts) (*Asset, error)
	UpdateAsset(UpdateAssetOpts) error
	GetAsset(urnString string) (*Asset, error)
	DeleteAsset(urnString string) error
	FindAssets(typeURN string, filter map[string]interface{}) ([]Asset, error)
}

var _ iService = (*Service)(nil)

// AssetTypeStore ...
type AssetTypeStore interface {
	CreateAssetType(*AssetType) error
	GetAssetType(urn AssetURN) (*AssetType, error)
}

// AssetStore ...
type AssetStore interface {
	CreateAsset(*Asset) error
	UpdateAsset(*Asset) error
	GetAsset(urn AssetURN) (*Asset, error)
	DeleteAsset(urn AssetURN) error
	FindFilter(typeURN AssetURN, filter map[string]interface{}) ([]Asset, error)
}

// Service ...
type Service struct {
	atStore    AssetTypeStore
	aStore     AssetStore
	pipelineID string

	assetTypes []string
	router     chi.Router
}

// Opts ...
type Opts struct {
	AssetStore     AssetStore
	AssetTypeStore AssetTypeStore
	PipelineID     string
}

func New(opts Opts) *Service {
	svc := &Service{
		aStore:     opts.AssetStore,
		atStore:    opts.AssetTypeStore,
		pipelineID: opts.PipelineID,
		router:     chi.NewRouter(),
	}
	svc.setupRoutes()

	return svc
}

// RegisterAssetTypeOpts ...
type RegisterAssetTypeOpts struct {
	Name       string
	PipelineID string // Shouldn't the service specify the PipelineID
	Labels     []string
	Version    int
	Schema     json.RawMessage
}

func (svc *Service) RegisterAssetType(opts RegisterAssetTypeOpts) error {
	at, err := newAssetType(newAssetTypeOpts{
		Name:       opts.Name,
		PipelineID: opts.PipelineID,
		Labels:     []string{},
		Version:    opts.Version,
		Schema:     opts.Schema,
	})

	if err != nil {
		return fmt.Errorf("could not create asset type: %w", err)
	}

	// Assert the asset type
	existingAT, err := svc.atStore.GetAssetType(at.URN())
	if err != nil && !errors.Is(err, ErrAssetTypeNotFound) {
		return err
	}
	if errors.Is(err, ErrAssetTypeNotFound) {
		return svc.atStore.CreateAssetType(at)
	}
	if !existingAT.Equals(at) {
		return ErrExistingAssetTypeMismatch
	}

	svc.assetTypes = append(svc.assetTypes, at.Name)

	return nil
}

func (svc *Service) GetAssetType(urnString string) (*AssetType, error) {
	urn, err := ParseAssetURN(urnString)
	if err != nil {
		return nil, err
	}

	return svc.atStore.GetAssetType(urn)
}

func (svc *Service) ListAssetTypes() ([]string, error) {
	return svc.assetTypes, nil
}

type CreateAssetOpts struct {
	Type    string
	Content json.RawMessage
}

func (svc *Service) CreateAsset(opts CreateAssetOpts) (*Asset, error) {
	at, err := svc.GetAssetType(opts.Type)
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
	fmt.Printf("urnString: %v\n", urnString)
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
func (svc *Service) FindAssets(typeURN string, filter map[string]interface{}) ([]Asset, error) {
	urn, err := ParseAssetURN(typeURN)
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
