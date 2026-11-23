package mongo

import (
	"context"
	"fmt"

	"github.com/sinJobsa/goWallet/gw-notification/internal/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionName = "large_transfers"

type Storage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func New(ctx context.Context, cfg config.MongoConfig) (*Storage, error) {
	clientOpts := options.Client().ApplyURI(cfg.URI)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	collection := client.Database(cfg.Database).Collection(collectionName)

	if err := ensureIndexes(ctx, collection); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ensure indexes: %w", err)
	}

	return &Storage{
		client:     client,
		collection: collection,
	}, nil
}

func ensureIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "timestamp", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (s *Storage) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
