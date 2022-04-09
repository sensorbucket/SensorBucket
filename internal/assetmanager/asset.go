package assetmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/xeipuuv/gojsonschema"
)

var (
	ErrValidationFailed = errors.New("validation failed")
)

// AssetDefinition ...
type AssetDefinition struct {
	Name       string
	PipelineID string
	Labels     []string
	Version    int
	Schema     json.RawMessage

	schema     *gojsonschema.Schema
	schemaSync sync.Once
}

// newAssetDefinitionOpts ...
type newAssetDefinitionOpts struct {
	Name       string
	PipelineID string
	Labels     []string
	Version    int
	Schema     json.RawMessage
}

func newAssetDefinition(opts newAssetDefinitionOpts) (*AssetDefinition, error) {
	at := &AssetDefinition{
		Name:       opts.Name,
		PipelineID: opts.PipelineID,
		Labels:     opts.Labels,
		Version:    opts.Version,
		Schema:     opts.Schema,

		schemaSync: sync.Once{},
	}

	// Verify schema
	loader := gojsonschema.NewBytesLoader(at.Schema)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return nil, fmt.Errorf("failed to compile asset definition schema: %w", err)
	}

	at.schema = schema

	return at, nil
}

func (at *AssetDefinition) URN() AssetURN {
	return AssetURN{
		PipelineID:      at.PipelineID,
		AssetDefinition: at.Name,
	}
}

func (at *AssetDefinition) Validate(c json.RawMessage) error {
	var err error

	// Parse schema only once and then store results
	at.schemaSync.Do(func() {
		schemaLoader := gojsonschema.NewStringLoader(string(at.Schema))
		at.schema, err = gojsonschema.NewSchema(schemaLoader)
	})
	if err != nil {
		return fmt.Errorf("%w: failed to parse asset schema for (%s): %s", ErrValidationFailed, at.URN(), err)
	}

	// Validate the content
	contentLoader := gojsonschema.NewBytesLoader(c)
	result, err := at.schema.Validate(contentLoader)
	if err != nil {
		return fmt.Errorf("%w: failed to validate asset content: %s", ErrValidationFailed, err)
	}

	// Create readable error message if failed
	if !result.Valid() {
		validationErrs := result.Errors()
		errorStrings := make([]string, 0, len(validationErrs))

		for _, err := range validationErrs {
			errorStrings = append(errorStrings,
				err.Field()+": "+err.Description(),
			)
		}
		return fmt.Errorf("%w: (%v)", ErrValidationFailed, strings.Join(errorStrings, ", "))
	}

	return nil
}

func (at *AssetDefinition) Equals(other *AssetDefinition) bool {
	return at.Name == other.Name && at.PipelineID == other.PipelineID && at.Version == other.Version
}

// Asset ...
type Asset struct {
	id      string
	at      *AssetDefinition
	Content json.RawMessage
}

func newAsset(at *AssetDefinition, content json.RawMessage) (*Asset, error) {
	if at == nil {
		return nil, errors.New("asset definition cannot be nil when creating newAsset")
	}

	return &Asset{
		id:      randomString(ASSET_ID_LENGTH),
		at:      at,
		Content: content,
	}, nil
}

func (a *Asset) URN() AssetURN {
	urn := a.at.URN()
	urn.AssetID = a.id
	return urn
}

func (a *Asset) Validate() error {
	return a.at.Validate(a.Content)
}
