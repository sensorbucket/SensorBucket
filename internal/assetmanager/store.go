package assetmanager

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	_ AssetStore = (*MongoDBStore)(nil)

	ASSET_COLLECTION      = "assets"
	ASSET_TYPE_COLLECTION = "asset_types"
)

// MongoDBStore ...
type MongoDBStore struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDBStore() *MongoDBStore {
	return &MongoDBStore{}
}

func (s *MongoDBStore) Connect(uri, database string) error {
	var err error

	s.client, err = mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to create mongodb client: %w", err)
	}

	if err := s.client.Connect(context.Background()); err != nil {
		return fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	s.db = s.client.Database(database)

	if err := s.migrate(); err != nil {
		return err
	}

	return nil
}

func (s *MongoDBStore) CreateAsset(asset *Asset) error {
	col := s.db.Collection(ASSET_COLLECTION)

	var model assetModel
	if err := model.From(asset); err != nil {
		return fmt.Errorf("could not create asset model: %w", err)
	}

	if _, err := col.InsertOne(context.Background(), model); err != nil {
		return fmt.Errorf("could not insert asset: %w", err)
	}

	return nil
}

func (s *MongoDBStore) UpdateAsset(asset *Asset) error {
	col := s.db.Collection(ASSET_COLLECTION)

	var model assetModel
	if err := model.From(asset); err != nil {
		return fmt.Errorf("could not create asset model: %w", err)
	}

	if _, err := col.UpdateOne(context.Background(), bson.D{{"type_urn", model.TypeURN}}, model); err != nil {
		return fmt.Errorf("could not update asset: %w", err)
	}

	return nil
}

func (s *MongoDBStore) GetAsset(urn AssetURN) (*Asset, error) {
	col := s.db.Collection(ASSET_COLLECTION)

	result := col.FindOne(context.Background(), bson.M{"urn": urn.String()})
	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to find asset: %w", err)
	}

	var model assetModel
	if err := result.Decode(&model); err != nil {
		return nil, fmt.Errorf("could not decode asset model: %w", err)
	}

	// Fetch corresponding asset type
	atURN, err := ParseAssetURN(model.TypeURN)
	atURN.AssetID = "" // We don't need the asset ID since we're referencing the asset type
	atModel, err := s.getAssetType(atURN)
	if err != nil {
		return nil, fmt.Errorf("could not get asset type: %w", err)
	}
	model.at = atModel

	asset, err := model.To()
	if err != nil {
		return nil, fmt.Errorf("could not convert db model to asset: %w", err)
	}

	return asset, nil
}

func (s *MongoDBStore) DeleteAsset(urn AssetURN) error {
	col := s.db.Collection(ASSET_COLLECTION)

	if _, err := col.DeleteOne(context.Background(), bson.M{"urn": urn.String()}); err != nil {
		return fmt.Errorf("could not delete asset: %w", err)
	}

	return nil
}

func (s *MongoDBStore) CreateAssetType(at *AssetType) error {
	col := s.db.Collection(ASSET_TYPE_COLLECTION)

	var model assetTypeModel
	if err := model.From(at); err != nil {
		return fmt.Errorf("could not create asset type model: %w", err)
	}

	if _, err := col.InsertOne(context.Background(), model); err != nil {
		return fmt.Errorf("could not insert asset type: %w", err)
	}

	return nil
}

func (s *MongoDBStore) getAssetType(urn AssetURN) (*assetTypeModel, error) {
	col := s.db.Collection(ASSET_TYPE_COLLECTION)

	result := col.FindOne(context.Background(), bson.M{"urn": urn.String()})
	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAssetTypeNotFound
		}
		return nil, fmt.Errorf("failed to find asset type: %w", err)
	}

	var model assetTypeModel
	if err := result.Decode(&model); err != nil {
		return nil, fmt.Errorf("could not decode asset type model: %w", err)
	}

	return &model, nil
}

func (s *MongoDBStore) GetAssetType(urn AssetURN) (*AssetType, error) {
	model, err := s.getAssetType(urn)
	if err != nil {
		return nil, err
	}

	at, err := model.To()
	if err != nil {
		return nil, fmt.Errorf("could not convert db model to asset type: %w", err)
	}

	return at, nil
}

func (s *MongoDBStore) migrate() error {
	if err := s.assertCollection(ASSET_COLLECTION); err != nil {
		return fmt.Errorf("could not assert asset collection: %w", err)
	}
	if err := s.assertCollection(ASSET_TYPE_COLLECTION); err != nil {
		return fmt.Errorf("could not assert asset types collection: %w", err)
	}

	// Assert indices
	if err := s.assertIndex(ASSET_COLLECTION, "urn"); err != nil {
		return fmt.Errorf("could not assert index on asset collection: %w", err)
	}
	if err := s.assertIndex(ASSET_TYPE_COLLECTION, "urn"); err != nil {
		return fmt.Errorf("could not assert index on asset types collection: %w", err)
	}

	return nil
}
