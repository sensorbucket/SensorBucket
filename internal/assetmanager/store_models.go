package assetmanager

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// assetDefinitionModel ...
type assetDefinitionModel struct {
	URN        string   `bson:"urn"`
	Labels     []string `bson:"labels"`
	Version    int      `bson:"version"`
	PrimaryKey string   `bson:"primary_key"`
	Schema     bson.Raw `bson:"schema"`
}

func (model *assetDefinitionModel) From(assetDefinition *AssetDefinition) error {
	// MongoDB works with bson, so convert JSON to bson
	var bsonSchema bson.Raw
	if err := bson.UnmarshalExtJSON(assetDefinition.Schema, true, &bsonSchema); err != nil {
		return err
	}

	newModel := assetDefinitionModel{
		URN:        assetDefinition.URN().String(),
		Labels:     assetDefinition.Labels,
		Version:    assetDefinition.Version,
		PrimaryKey: assetDefinition.PrimaryKey,
		Schema:     bsonSchema,
	}
	*model = newModel

	return nil
}

func (model *assetDefinitionModel) To() (*AssetDefinition, error) {
	// MongoDB works with bson, so convert bson to JSON
	jsonSchema, err := bson.MarshalExtJSON(model.Schema, true, false)
	if err != nil {
		return nil, err
	}

	// Database only stores the URN, we'll have to decode it
	urn, err := ParseAssetURN(model.URN)
	if err != nil {
		return nil, err
	}

	at, err := newAssetDefinition(newAssetDefinitionOpts{
		Name:       urn.AssetDefinition,
		PipelineID: urn.PipelineID,
		Labels:     model.Labels,
		Version:    model.Version,
		PrimaryKey: model.PrimaryKey,
		Schema:     jsonSchema,
	})
	if err != nil {
		return nil, err
	}

	return at, nil
}

// assetModel ...
type assetModel struct {
	URN           string                `bson:"urn"`
	DefinitionURN string                `bson:"definition_urn"`
	Content       bson.Raw              `bson:"content"`
	at            *assetDefinitionModel `bson:"at"` // Not stored in database, should be joined
}

func (model *assetModel) From(asset *Asset) error {
	// MongoDB works with bson, so convert JSON to bson
	var bsonContent bson.Raw
	if err := bson.UnmarshalExtJSON(asset.Content, false, &bsonContent); err != nil {
		return fmt.Errorf("could not unmarshal asset content to bson: %w", err)
	}

	var atModel assetDefinitionModel
	if err := atModel.From(asset.Definition); err != nil {
		return fmt.Errorf("could not convert asset definition to model: %w", err)
	}

	newModel := assetModel{
		URN:           asset.URN().String(),
		DefinitionURN: asset.Definition.URN().String(),
		Content:       bsonContent,
		at:            &atModel,
	}
	*model = newModel

	return nil
}

func (model *assetModel) To() (*Asset, error) {
	// Convert asset definition first
	if model.at == nil {
		return nil, fmt.Errorf("asset definition in asset model is nil")
	}
	at, err := model.at.To()
	if err != nil {
		return nil, err
	}

	// MongoDB works with bson, so convert bson to JSON
	jsonContent, err := bson.MarshalExtJSON(model.Content, false, false)
	if err != nil {
		return nil, err
	}

	// Database only stores the URN, we'll have to decode it
	urn, err := ParseAssetURN(model.URN)
	if err != nil {
		return nil, err
	}

	asset := &Asset{
		ID:         urn.AssetID,
		Definition: at,
		Content:    jsonContent,
	}

	if err := asset.Validate(); err != nil {
		return nil, err
	}

	return asset, nil
}
