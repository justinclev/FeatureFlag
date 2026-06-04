package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/featureflags/feature-api/internal/models"
)

// FlagRepository defines the storage contract for feature flags.
// Handlers depend on this interface, not on concrete infrastructure.
type FlagRepository interface {
	List(ctx context.Context, limit, offset int64) ([]models.Flag, error)
	GetByID(ctx context.Context, id string) (*models.Flag, error)
	GetByKey(ctx context.Context, key string) (*models.Flag, error)
	Create(ctx context.Context, req models.CreateFlagRequest) (*models.Flag, error)
	Update(ctx context.Context, id string, req models.UpdateFlagRequest) (*models.Flag, error)
	Delete(ctx context.Context, id string) error
	Ready(ctx context.Context) error
}

// RedisClient defines the subset of redis.Client methods used by the repository.
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
}

// MongoCollection defines the subset of mongo.Collection methods used by the repository.
type MongoCollection interface {
	Find(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error)
	FindOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOneOptions]) *mongo.SingleResult
	InsertOne(ctx context.Context, document interface{}, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error)
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult
	DeleteOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error)
	CountDocuments(ctx context.Context, filter interface{}, opts ...options.Lister[options.CountOptions]) (int64, error)
	Database() *mongo.Database
}
