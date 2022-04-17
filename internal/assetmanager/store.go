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

	ASSET_COLLECTION            = "assets"
	ASSET_DEFINITION_COLLECTION = "asset_definitions"
)

// MongoDBStore implements the AssetStore interface using a MongoDB backend.
type MongoDBStore struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewMongoDBStore creates a new MongoDBStore instance.
func NewMongoDBStore() *MongoDBStore {
	return &MongoDBStore{}
}

// Connect connects to the MongoDB backend with the given database and returns an error if the connection fails.
// Will also migrate the database if possible.
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

	if _, err := col.UpdateOne(context.Background(), bson.D{{"urn", model.URN}}, bson.D{{"$set", bson.D{{"content", model.Content}}}}); err != nil {
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

	// Fetch corresponding asset definition
	atURN, err := ParseAssetURN(model.DefinitionURN)
	if err != nil {
		return nil, fmt.Errorf("could not parse asset definition urn: %w", err)
	}
	atURN.AssetID = "" // We don't need the asset ID since we're referencing the asset definition
	atModel, err := s.getAssetDefinition(atURN)
	if err != nil {
		return nil, fmt.Errorf("could not get asset definition: %w", err)
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

func (s *MongoDBStore) FindFilter(definitionURN AssetURN, cf map[string]interface{}) ([]Asset, error) {
	col := s.db.Collection(ASSET_COLLECTION)

	// Create mongo query filter
	filter := bson.M{
		"definition_urn": definitionURN.String(),
	}
	for key, value := range cf {
		filter["content."+key] = value
	}

	cursor, err := col.Find(context.Background(), filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("error finding asset: %w", err)
	}

	// Fetch corresponding asset definition
	atModel, err := s.getAssetDefinition(definitionURN)
	if err != nil {
		return nil, fmt.Errorf("could not get asset definition: %w", err)
	}

	// Decode all models and convert to asset
	var models []assetModel
	if err := cursor.All(context.Background(), &models); err != nil {
		return nil, fmt.Errorf("error decoding asset models: %w", err)
	}

	var assets = make([]Asset, len(models))
	for ix, model := range models {
		model.at = atModel

		asset, err := model.To()
		if err != nil {
			return nil, fmt.Errorf("could not convert db model to asset: %w", err)
		}

		assets[ix] = *asset
	}

	return assets, nil
}

func (s *MongoDBStore) CreateAssetDefinition(at *AssetDefinition) error {
	col := s.db.Collection(ASSET_DEFINITION_COLLECTION)

	var model assetDefinitionModel
	if err := model.From(at); err != nil {
		return fmt.Errorf("could not create asset definition model: %w", err)
	}

	if _, err := col.InsertOne(context.Background(), model); err != nil {
		return fmt.Errorf("could not insert asset definition: %w", err)
	}

	return nil
}

func (s *MongoDBStore) getAssetDefinition(urn AssetURN) (*assetDefinitionModel, error) {
	col := s.db.Collection(ASSET_DEFINITION_COLLECTION)

	result := col.FindOne(context.Background(), bson.M{"urn": urn.String()})
	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAssetDefinitionNotFound
		}
		return nil, fmt.Errorf("failed to find asset definition: %w", err)
	}

	var model assetDefinitionModel
	if err := result.Decode(&model); err != nil {
		return nil, fmt.Errorf("could not decode asset definition model: %w", err)
	}

	return &model, nil
}

func (s *MongoDBStore) GetAssetDefinition(urn AssetURN) (*AssetDefinition, error) {
	model, err := s.getAssetDefinition(urn)
	if err != nil {
		return nil, err
	}

	at, err := model.To()
	if err != nil {
		return nil, fmt.Errorf("could not convert db model to asset definition: %w", err)
	}

	return at, nil
}

func (s *MongoDBStore) migrate() error {
	if err := s.assertCollection(ASSET_COLLECTION); err != nil {
		return fmt.Errorf("could not assert asset collection: %w", err)
	}
	if err := s.assertCollection(ASSET_DEFINITION_COLLECTION); err != nil {
		return fmt.Errorf("could not assert asset definitions collection: %w", err)
	}

	// Assert indices
	if err := s.assertIndex(ASSET_COLLECTION, "urn"); err != nil {
		return fmt.Errorf("could not assert index on asset collection: %w", err)
	}
	if err := s.assertIndex(ASSET_DEFINITION_COLLECTION, "urn"); err != nil {
		return fmt.Errorf("could not assert index on asset definitions collection: %w", err)
	}

	return nil
}
