package assetmanager

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (store *MongoDBStore) assertCollection(collectionName string) error {
	exists, err := store.collectionExists(collectionName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists: %w", err)
	}

	if !exists {
		return store.createCollection(collectionName)
	}

	return nil
}

func (store *MongoDBStore) assertIndex(collection, key string) error {
	indexes := store.db.Collection(collection).Indexes()

	opts := mongo.IndexModel{
		Keys: bson.D{{key, 1}},
	}

	if _, err := indexes.CreateOne(context.Background(), opts); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

func (store *MongoDBStore) createCollection(name string) error {
	if err := store.db.CreateCollection(context.Background(), name); err != nil {
		return fmt.Errorf("failed to create collection %T: %w ", err, err)
	}
	return nil
}

func (store *MongoDBStore) collectionExists(name string) (bool, error) {
	collections, err := store.db.ListCollectionNames(context.Background(), bson.D{{"name", name}})
	if err != nil {
		return false, fmt.Errorf("failed to list collections: %w", err)
	}

	return len(collections) > 0, nil
}
