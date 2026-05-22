package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/featureflags/feature-api/internal/models"
)

const (
	flagCachePrefix = "flags:id:"
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// MongoRedisRepository implements FlagRepository using MongoDB for persistence
// and Redis as a read-through cache.
type MongoRedisRepository struct {
	col      *mongo.Collection
	rdb      RedisClient
	cacheTTL time.Duration
}

// NewMongoRedisRepository constructs a MongoRedisRepository.
func NewMongoRedisRepository(col *mongo.Collection, rdb RedisClient, cacheTTL time.Duration) *MongoRedisRepository {
	return &MongoRedisRepository{col: col, rdb: rdb, cacheTTL: cacheTTL}
}

func (r *MongoRedisRepository) List(ctx context.Context) ([]models.Flag, error) {
	cursor, err := r.col.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find flags: %w", err)
	}
	defer cursor.Close(ctx)

	var flags []models.Flag
	if err := cursor.All(ctx, &flags); err != nil {
		return nil, fmt.Errorf("decode flags: %w", err)
	}
	if flags == nil {
		flags = []models.Flag{}
	}
	return flags, nil
}

func (r *MongoRedisRepository) GetByID(ctx context.Context, id string) (*models.Flag, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	cacheKey := flagCachePrefix + id
	if cached, err := r.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var flag models.Flag
		if jsonErr := json.Unmarshal(cached, &flag); jsonErr == nil {
			return &flag, nil
		}
	}

	var flag models.Flag
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&flag); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find flag: %w", err)
	}

	if payload, err := json.Marshal(flag); err == nil {
		_ = r.rdb.Set(ctx, cacheKey, payload, r.cacheTTL).Err()
	}
	return &flag, nil
}

func (r *MongoRedisRepository) Create(ctx context.Context, req models.CreateFlagRequest) (*models.Flag, error) {
	now := time.Now().UTC()
	flag := models.Flag{
		ID:           bson.NewObjectID(),
		Name:         req.Name,
		Key:          req.Key,
		Enabled:      req.Enabled,
		Description:  req.Description,
		DefaultValue: req.DefaultValue,
		Rules:        req.Rules,
		CreatedBy:    req.CreatedBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if flag.Rules == nil {
		flag.Rules = []models.Rule{}
	}

	if _, err := r.col.InsertOne(ctx, flag); err != nil {
		return nil, fmt.Errorf("insert flag: %w", err)
	}

	r.invalidate(ctx, flag.ID.Hex())
	return &flag, nil
}

func (r *MongoRedisRepository) Update(ctx context.Context, id string, req models.UpdateFlagRequest) (*models.Flag, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	fields := bson.M{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Key != nil {
		fields["key"] = *req.Key
	}
	if req.Enabled != nil {
		fields["enabled"] = *req.Enabled
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.DefaultValue != nil {
		fields["defaultValue"] = *req.DefaultValue
	}
	if req.Rules != nil {
		fields["rules"] = *req.Rules
	}
	if len(fields) == 0 {
		return nil, ErrNoFields
	}
	fields["updatedAt"] = time.Now().UTC()
	if req.UpdatedBy != "" {
		fields["updatedBy"] = req.UpdatedBy
	}

	result, err := r.col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": fields})
	if err != nil {
		return nil, fmt.Errorf("update flag: %w", err)
	}
	if result.MatchedCount == 0 {
		return nil, ErrNotFound
	}

	r.invalidate(ctx, id)

	var flag models.Flag
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&flag); err != nil {
		return nil, fmt.Errorf("fetch updated flag: %w", err)
	}
	return &flag, nil
}

func (r *MongoRedisRepository) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	result, err := r.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("delete flag: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	r.invalidate(ctx, id)
	return nil
}

func (r *MongoRedisRepository) invalidate(ctx context.Context, id string) {
	_ = r.rdb.Del(ctx, flagCachePrefix+id).Err()
}
