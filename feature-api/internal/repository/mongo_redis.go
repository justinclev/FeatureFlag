package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/sync/singleflight"

	"github.com/featureflags/feature-api/internal/models"
)

// RedisClient defines the subset of redis.Client methods used by the repository.
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// MongoCollection defines the subset of mongo.Collection methods used by the repository.
type MongoCollection interface {
	Find(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error)
	FindOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOneOptions]) *mongo.SingleResult
	InsertOne(ctx context.Context, document interface{}, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error)
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult
	DeleteOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error)
}

// MongoRedisRepository implements FlagRepository using MongoDB for persistence
// and a multi-tier cache (L1 LRU, L2 Redis).
type MongoRedisRepository struct {
	col         MongoCollection
	rdb         RedisClient
	cacheTTL    time.Duration
	cachePrefix string
	sf          singleflight.Group
	l1          *expirable.LRU[string, *models.Flag]
}

// NewMongoRedisRepository constructs a MongoRedisRepository.
func NewMongoRedisRepository(col MongoCollection, rdb RedisClient, cacheTTL time.Duration, cachePrefix string) *MongoRedisRepository {
	// Initialize L1 cache with 1000 items and a 10-second TTL.
	// This ensures that even in a multi-node environment, stale data in memory
	// is purged automatically within 10 seconds.
	l1 := expirable.NewLRU[string, *models.Flag](1000, nil, 10*time.Second)
	return &MongoRedisRepository{
		col:         col,
		rdb:         rdb,
		cacheTTL:    cacheTTL,
		cachePrefix: cachePrefix,
		sf:          singleflight.Group{},
		l1:          l1,
	}
}

// List returns a page of feature flags from the database.
func (r *MongoRedisRepository) List(ctx context.Context, limit, offset int64) ([]models.Flag, error) {
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	opts.SetSkip(offset)

	cursor, err := r.col.Find(ctx, bson.D{}, opts)
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

// GetByID retrieves a single feature flag by its ID, checking L1 and L2 cache first.
func (r *MongoRedisRepository) GetByID(ctx context.Context, id string) (*models.Flag, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	// Tier 1: L1 In-Memory LRU (Near Zero Cost)
	if flag, ok := r.l1.Get(id); ok {
		return flag, nil
	}

	cacheKey := r.cachePrefix + id
	// Tier 2: L2 Redis Cache (Network + JSON Overhead)
	if cached, err := r.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var flag models.Flag
		if jsonErr := json.Unmarshal(cached, &flag); jsonErr == nil {
			r.l1.Add(id, &flag) // Promote to L1
			return &flag, nil
		}
	}

	// Tier 3: Singleflight to DB (Prevents Thundering Herd)
	val, err, _ := r.sf.Do(id, func() (interface{}, error) {
		var flag models.Flag
		if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&flag); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("find flag: %w", err)
		}

		// Update Redis (L2)
		if payload, err := json.Marshal(flag); err == nil {
			_ = r.rdb.Set(ctx, cacheKey, payload, r.cacheTTL).Err()
		}
		// Update L1
		r.l1.Add(id, &flag)

		return &flag, nil
	})

	if err != nil {
		return nil, err
	}
	return val.(*models.Flag), nil
}

// Create inserts a new feature flag into the database.
func (r *MongoRedisRepository) Create(ctx context.Context, req models.CreateFlagRequest) (*models.Flag, error) {
	now := time.Now().UTC()
	strategy := req.RuleMatchStrategy
	if strategy == "" {
		strategy = models.RuleMatchStrategyAny
	}

	flag := models.Flag{
		ID:                bson.NewObjectID(),
		Name:              req.Name,
		Key:               req.Key,
		Enabled:           req.Enabled,
		Description:       req.Description,
		DefaultValue:      req.DefaultValue,
		Rules:             req.Rules,
		RuleMatchStrategy: strategy,
		CreatedBy:         req.CreatedBy,
		CreatedAt:         now,
		UpdatedAt:         now,
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

// Update modifies an existing feature flag and invalidates all cache tiers.
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
	if req.RuleMatchStrategy != nil {
		fields["ruleMatchStrategy"] = *req.RuleMatchStrategy
	}
	if len(fields) == 0 {
		return nil, ErrNoFields
	}
	fields["updatedAt"] = time.Now().UTC()
	if req.UpdatedBy != "" {
		fields["updatedBy"] = req.UpdatedBy
	}

	var flag models.Flag
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = r.col.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": fields}, opts).Decode(&flag)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update flag: %w", err)
	}

	r.invalidate(ctx, id)
	return &flag, nil
}

// Delete removes a feature flag from the database and invalidates all cache tiers.
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
	r.l1.Remove(id)                            // Clear L1
	_ = r.rdb.Del(ctx, r.cachePrefix+id).Err() // Clear L2
}
