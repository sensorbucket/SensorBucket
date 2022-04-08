package assetmanager

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// assetTypeModel ...
type assetTypeModel struct {
	URN     string   `bson:"urn"`
	Labels  []string `bson:"labels"`
	Version int      `bson:"version"`
	Schema  bson.Raw `bson:"schema"`
}

func (model *assetTypeModel) From(assetType *AssetType) error {
	// MongoDB works with bson, so convert JSON to bson
	var bsonSchema bson.Raw
	if err := bson.UnmarshalExtJSON(assetType.Schema, true, &bsonSchema); err != nil {
		return err
	}

	newModel := assetTypeModel{
		URN:     assetType.URN().String(),
		Labels:  assetType.Labels,
		Version: assetType.Version,
		Schema:  bsonSchema,
	}
	*model = newModel

	return nil
}

func (model *assetTypeModel) To() (*AssetType, error) {
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

	at, err := newAssetType(newAssetTypeOpts{
		Name:       urn.AssetType,
		PipelineID: urn.PipelineID,
		Labels:     model.Labels,
		Version:    model.Version,
		Schema:     jsonSchema,
	})
	if err != nil {
		return nil, err
	}

	return at, nil
}

// assetModel ...
type assetModel struct {
	URN     string          `bson:"urn"`
	TypeURN string          `bson:"type_urn"`
	Content bson.Raw        `bson:"content"`
	at      *assetTypeModel `bson:"at"` // Not stored in database, should be joined
}

func (model *assetModel) From(asset *Asset) error {
	// MongoDB works with bson, so convert JSON to bson
	var bsonContent bson.Raw
	if err := bson.UnmarshalExtJSON(asset.Content, false, &bsonContent); err != nil {
		return fmt.Errorf("could not unmarshal asset content to bson: %w", err)
	}

	var atModel assetTypeModel
	if err := atModel.From(asset.at); err != nil {
		return fmt.Errorf("could not convert asset type to model: %w", err)
	}

	newModel := assetModel{
		URN:     asset.URN().String(),
		TypeURN: asset.at.URN().String(),
		Content: bsonContent,
		at:      &atModel,
	}
	*model = newModel

	return nil
}

func (model *assetModel) To() (*Asset, error) {
	// Convert asset type first
	at, err := model.at.To()
	if err != nil {
		return nil, err
	}

	// MongoDB works with bson, so convert bson to JSON
	jsonContent, err := bson.MarshalExtJSON(model.Content, true, false)
	if err != nil {
		return nil, err
	}

	// Database only stores the URN, we'll have to decode it
	urn, err := ParseAssetURN(model.URN)
	if err != nil {
		return nil, err
	}

	asset := &Asset{
		id:      urn.AssetID,
		at:      at,
		Content: jsonContent,
	}

	if err := asset.Validate(); err != nil {
		return nil, err
	}

	return asset, nil
}
