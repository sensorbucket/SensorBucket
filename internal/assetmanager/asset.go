package assetmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"
)

const (
	ILLEGAL_PK_CHARACTERS = " #:"
)

var (
	ErrValidationFailed = errors.New("validation failed")
)

// AssetDefinition contains the asset structure, such as schema, labels, keys, etc.
type AssetDefinition struct {
	Name       string
	PipelineID string
	PrimaryKey string
	Labels     []string
	Version    int
	Schema     json.RawMessage

	schema     *gojsonschema.Schema
	schemaSync sync.Once
}

type newAssetDefinitionOpts struct {
	Name       string
	PipelineID string
	Labels     []string
	Version    int
	PrimaryKey string
	Schema     json.RawMessage
}

func newAssetDefinition(opts newAssetDefinitionOpts) (*AssetDefinition, error) {
	at := &AssetDefinition{
		Name:       opts.Name,
		PipelineID: opts.PipelineID,
		PrimaryKey: opts.PrimaryKey,
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

	// Validate PrimaryKey
	if at.PrimaryKey != "" {
		if err := validatePrimaryKey(at.PrimaryKey); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrValidationFailed, err)
		}
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

// Validate checks if the given json content adheres to this asset definition
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

// Asset contains the asset content in json, and references its definition
type Asset struct {
	ID         string
	Definition *AssetDefinition
	Content    json.RawMessage
}

func newAsset(at *AssetDefinition, content json.RawMessage) (*Asset, error) {
	var err error

	if at == nil {
		return nil, errors.New("asset definition cannot be nil when creating newAsset")
	}

	// Use primary key as asset id if available
	var id string
	if at.PrimaryKey != "" {
		id, err = generateIDFromPK(at.PrimaryKey, content)
		if err != nil {
			return nil, fmt.Errorf("failed to generate asset ID: %w", err)
		}
	}

	if id == "" {
		id = randomString(ASSET_ID_LENGTH)
	}

	return &Asset{
		ID:         id,
		Definition: at,
		Content:    content,
	}, nil
}

func (a *Asset) URN() AssetURN {
	urn := a.Definition.URN()
	urn.AssetID = a.ID
	return urn
}

func (a *Asset) Validate() error {
	return a.Definition.Validate(a.Content)
}

var R_PK_FIELD = regexp.MustCompile(`(?miU)(\${.+})`)

func generateIDFromPK(pk string, content json.RawMessage) (string, error) {
	keys := R_PK_FIELD.FindAllString(pk, -1)
	if len(keys) == 0 {
		return "", errors.New("primary key must contain atleast one field")
	}

	for _, key := range keys {
		value := gjson.GetBytes(content, strings.Trim(key, "${}"))
		if !value.Exists() {
			return "", fmt.Errorf("primary key field (%s) does not exist in content", key)
		}
		pk = strings.ReplaceAll(pk, key, value.String())
	}

	return pk, nil
}

func validatePrimaryKey(pk string) error {
	if strings.ContainsAny(pk, ILLEGAL_PK_CHARACTERS) {
		return errors.New("primary key contains illegal characters")
	}

	if !strings.HasPrefix(pk, "${") || !strings.HasSuffix(pk, "}") {
		return errors.New("primary key must have at least one template")
	}

	return nil
}
