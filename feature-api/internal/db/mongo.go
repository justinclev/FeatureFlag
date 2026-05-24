package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/featureflags/feature-api/internal/config"
)

// Connect opens a MongoDB connection and returns the client and target database.
// The caller owns the client lifecycle and must call Disconnect.
func Connect(cfg *config.Config) (*mongo.Client, *mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("mongo ping: %w", err)
	}

	db := client.Database(cfg.MongoDBName)

	// Ensure Indexes
	if err := ensureIndexes(ctx, db, cfg.MongoCollectionName); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("ensure indexes: %w", err)
	}

	return client, db, nil
}

func ensureIndexes(ctx context.Context, db *mongo.Database, collectionName string) error {
	col := db.Collection(collectionName)

	// Index on "key" (unique)
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "key", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// Index on "enabled" for common filtering
	_, err = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "enabled", Value: 1}},
	})
	return err
}

// Disconnect closes the MongoDB client gracefully.
func Disconnect(client *mongo.Client) {
	if client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = client.Disconnect(ctx)
}
