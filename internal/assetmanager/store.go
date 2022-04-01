package assetmanager

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	_ Store = (*MongoDBStore)(nil)
)

// assetModel ...
type assetModel struct {
	URN     string   `bson:"urn"`
	Content bson.Raw `bson:"content"`
}

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

	return nil
}

func (s *MongoDBStore) Register(asset AssetSchema) error {
	if err := assertAssetCollection(s.db, asset.Type); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	if err := assertAssetIndex(s.db, asset.Type); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

func (s *MongoDBStore) Create(asset *Asset) error {
	col := s.getCollection(asset.Type)

	model, err := assetToModel(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset to model: %w", err)
	}

	result, err := col.InsertOne(context.Background(), model)
	if err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}

	// TODO: is this required? Should be use the ID?
	if result.InsertedID == nil {
		return fmt.Errorf("failed to create asset: no id returned")
	}

	return nil
}

func (s *MongoDBStore) Update(asset *Asset) error {
	col := s.getCollection(asset.Type)

	model, err := assetToModel(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset to model: %w", err)
	}

	if _, err := col.UpdateOne(context.Background(), bson.D{{"urn", asset.URN}}, model); err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	return nil
}

func (s *MongoDBStore) Get(urn AssetURN) (*Asset, error) {
	col := s.getCollection(urn.AssetType)

	result := col.FindOne(context.Background(), bson.D{{"urn", urn.String()}})
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	var asset Asset
	result.Decode(&asset)

	return &asset, nil
}

func (s *MongoDBStore) Delete(urn AssetURN) error {
	col := s.getCollection(urn.AssetType)

	if _, err := col.DeleteOne(context.Background(), bson.D{{"urn", urn.String()}}); err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	return nil
}

//
// Helper functions
//

func assetToModel(asset *Asset) (*assetModel, error) {
	var content bson.Raw
	if err := bson.UnmarshalExtJSON([]byte(asset.Content), true, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %w", err)
	}

	return &assetModel{
		URN:     asset.URN(),
		Content: content,
	}, nil
}

func createCollection(db *mongo.Database, name string) error {
	if err := db.CreateCollection(context.Background(), name); err != nil {
		return fmt.Errorf("failed to create collection %T: %w ", err, err)
	}
	return nil
}

func collectionExists(db *mongo.Database, name string) (bool, error) {
	collections, err := db.ListCollectionNames(context.Background(), bson.D{{"name", name}})
	if err != nil {
		return false, fmt.Errorf("failed to list collections: %w", err)
	}

	return len(collections) > 0, nil
}

func assertAssetCollection(db *mongo.Database, assetType string) error {
	collectionName := createAssetCollectionName(assetType)
	exists, err := collectionExists(db, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists: %w", err)
	}

	if !exists {
		return createCollection(db, collectionName)
	}

	return nil
}
func createIndex(db *mongo.Database, collection, key string) error {
	indexes := db.Collection(collection).Indexes()

	opts := mongo.IndexModel{
		Keys: bson.D{{key, 1}},
	}

	if _, err := indexes.CreateOne(context.Background(), opts); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

func assertAssetIndex(db *mongo.Database, assetType string) error {
	collectionName := createAssetCollectionName(assetType)
	return createIndex(db, collectionName, "urn")
}

func (s *MongoDBStore) getCollection(assetType string) *mongo.Collection {
	return s.db.Collection(createAssetCollectionName(assetType))
}

func createAssetCollectionName(assetType string) string {
	return "asset_" + assetType
}
