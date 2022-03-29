package assetmanager

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

var (
	ErrInvalidSchema          = errors.New("invalid schema")
	ErrAssetTypeNotRegistered = errors.New("asset type not registered")
	ErrValidationFailed       = errors.New("validation failed")
)

// SchemaRegistry ...
type SchemaRegistry struct {
	schemas map[string]*gojsonschema.Schema
}

func newSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		schemas: make(map[string]*gojsonschema.Schema),
	}
}

func (sr *SchemaRegistry) Register(asset AssetSchema) error {
	loader := gojsonschema.NewBytesLoader(asset.Schema)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSchema, err)
	}

	sr.schemas[asset.Type] = schema
	return nil
}

func (sr *SchemaRegistry) Validate(asset *Asset) error {
	loader := gojsonschema.NewBytesLoader(asset.Content)
	schema, ok := sr.schemas[asset.Type]
	if !ok {
		return ErrAssetTypeNotRegistered
	}

	result, err := schema.Validate(loader)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

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
