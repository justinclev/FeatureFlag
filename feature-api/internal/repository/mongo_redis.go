package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/sync/singleflight"

	"github.com/featureflags/feature-api/internal/models"
)

const (
	negCacheValue = "__404__"
	shardCount    = 64
)

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
	Database() *mongo.Database
}

type cacheItem struct {
	flag      *models.Flag
	expiresAt time.Time
}

type shard struct {
	sync.RWMutex
	data map[string]cacheItem
}

// ShardedL1Cache is a high-concurrency in-memory cache that minimizes lock contention.
type ShardedL1Cache struct {
	shards [shardCount]*shard
}

func newShardedL1Cache() *ShardedL1Cache {
	c := &ShardedL1Cache{}
	for i := 0; i < shardCount; i++ {
		c.shards[i] = &shard{data: make(map[string]cacheItem)}
	}
	go c.janitor()
	return c
}

func (c *ShardedL1Cache) janitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		for i := 0; i < shardCount; i++ {
			s := c.shards[i]
			s.Lock()
			for k, v := range s.data {
				if time.Now().After(v.expiresAt) {
					delete(s.data, k)
				}
			}
			s.Unlock()
		}
	}
}

func (c *ShardedL1Cache) getShard(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.shards[h.Sum32()%shardCount]
}

func (c *ShardedL1Cache) Get(key string) (*models.Flag, bool) {
	s := c.getShard(key)
	s.RLock()
	item, ok := s.data[key]
	s.RUnlock()

	if !ok || time.Now().After(item.expiresAt) {
		return nil, false
	}
	return item.flag, true
}

func (c *ShardedL1Cache) Set(key string, flag *models.Flag, ttl time.Duration) {
	s := c.getShard(key)
	s.Lock()
	s.data[key] = cacheItem{
		flag:      flag,
		expiresAt: time.Now().Add(ttl),
	}
	s.Unlock()
}

func (c *ShardedL1Cache) Remove(key string) {
	s := c.getShard(key)
	s.Lock()
	delete(s.data, key)
	s.Unlock()
}

// MongoRedisRepository implements FlagRepository using MongoDB for persistence
// and a multi-tier cache.
type MongoRedisRepository struct {
	col         MongoCollection
	rdb         RedisClient
	cacheTTL    time.Duration
	cachePrefix string
	sf          singleflight.Group
	l1          *ShardedL1Cache
}

// NewMongoRedisRepository constructs a MongoRedisRepository.
func NewMongoRedisRepository(col MongoCollection, rdb RedisClient, cacheTTL time.Duration, cachePrefix string) *MongoRedisRepository {
	return &MongoRedisRepository{
		col:         col,
		rdb:         rdb,
		cacheTTL:    cacheTTL,
		cachePrefix: cachePrefix,
		sf:          singleflight.Group{},
		l1:          newShardedL1Cache(),
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

// GetByID retrieves a single feature flag by its ID.
func (r *MongoRedisRepository) GetByID(ctx context.Context, id string) (*models.Flag, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	var flag models.Flag
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&flag); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find flag: %w", err)
	}
	return &flag, nil
}

// GetByKey retrieves a single feature flag by its unique key, checking caches first.
func (r *MongoRedisRepository) GetByKey(ctx context.Context, key string) (*models.Flag, error) {
	// Tier 1: Sharded L1 Cache
	if flag, ok := r.l1.Get(key); ok {
		if flag == nil {
			return nil, ErrNotFound
		}
		return flag.Clone(), nil
	}

	cacheKey := r.cachePrefix + key

	// Tier 2: L2 Redis Cache
	if cached, err := r.rdb.Get(ctx, cacheKey).Result(); err == nil {
		if cached == negCacheValue {
			r.l1.Set(key, nil, 1*time.Minute)
			return nil, ErrNotFound
		}
		var flag models.Flag
		if jsonErr := json.Unmarshal([]byte(cached), &flag); jsonErr == nil {
			r.l1.Set(key, &flag, r.cacheTTL)
			return &flag, nil
		}
	}

	// Tier 3: Singleflight to DB
	val, err, _ := r.sf.Do(key, func() (interface{}, error) {
		dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var flag models.Flag
		if err := r.col.FindOne(dbCtx, bson.M{"key": key}).Decode(&flag); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				_ = r.rdb.Set(dbCtx, cacheKey, negCacheValue, 1*time.Minute).Err()
				r.l1.Set(key, nil, 1*time.Minute)
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("find flag: %w", err)
		}

		if payload, err := json.Marshal(flag); err == nil {
			_ = r.rdb.Set(dbCtx, cacheKey, payload, r.cacheTTL).Err()
		}
		r.l1.Set(key, &flag, r.cacheTTL)

		return &flag, nil
	})

	if err != nil {
		return nil, err
	}
	return val.(*models.Flag).Clone(), nil
}

// Create inserts a new feature flag into the database.
func (r *MongoRedisRepository) Create(ctx context.Context, req models.CreateFlagRequest) (*models.Flag, error) {
	now := time.Now().UTC()
	flag := models.Flag{
		ID:                bson.NewObjectID(),
		Name:              req.Name,
		Key:               req.Key,
		Enabled:           req.Enabled,
		Description:       req.Description,
		DefaultValue:      req.DefaultValue,
		Rules:             req.Rules,
		RuleMatchStrategy: req.RuleMatchStrategy,
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

	r.invalidate(flag.Key)
	return &flag, nil
}

// Update modifies an existing feature flag and invalidates all cache tiers.
func (r *MongoRedisRepository) Update(ctx context.Context, id string, req models.UpdateFlagRequest) (*models.Flag, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
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

	r.invalidate(current.Key)
	if req.Key != nil && *req.Key != current.Key {
		r.invalidate(*req.Key)
	}
	return &flag, nil
}

// Delete removes a feature flag from the database and invalidates all cache tiers.
func (r *MongoRedisRepository) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	current, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	result, err := r.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("delete flag: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	r.invalidate(current.Key)
	return nil
}

func (r *MongoRedisRepository) invalidate(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r.l1.Remove(key)
	_ = r.rdb.Del(ctx, r.cachePrefix+key).Err()
}

// Ready verifies that both MongoDB and Redis are reachable.
func (r *MongoRedisRepository) Ready(ctx context.Context) error {
	if err := r.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis not ready: %w", err)
	}
	if err := r.col.Database().Client().Ping(ctx, nil); err != nil {
		return fmt.Errorf("mongodb not ready: %w", err)
	}
	return nil
}
