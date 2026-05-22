package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/featureflags/feature-api/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type fakeRedis struct {
	store map[string][]byte
}

func (f *fakeRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	if val, ok := f.store[key]; ok {
		return redis.NewStringResult(string(val), nil)
	}
	return redis.NewStringResult("", errors.New("not found"))
}
func (f *fakeRedis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	b, _ := json.Marshal(value)
	f.store[key] = b
	return redis.NewStatusResult("OK", nil)
}
func (f *fakeRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	for _, k := range keys {
		delete(f.store, k)
	}
	return redis.NewIntResult(int64(len(keys)), nil)
}

func TestMongoRedisRepository_List(t *testing.T) {
	ctx := context.Background()

	// 1. Spin up an ephemeral MongoDB container for the test
	mongodbContainer, err := mongodb.Run(ctx, "mongo:6")
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	defer mongodbContainer.Terminate(ctx)

	uri, _ := mongodbContainer.ConnectionString(ctx)
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("failed to connect to test container: %v", err)
	}
	defer client.Disconnect(ctx)

	col := client.Database("testdb").Collection("flags")
	col.InsertOne(ctx, models.Flag{ID: bson.NewObjectID(), Name: "flag1"})

	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second)

	flags, err := repo.List(ctx)
	if err != nil || len(flags) != 1 || flags[0].Name != "flag1" {
		t.Errorf("expected 1 flag, got %v, %v", flags, err)
	}
}

func TestFakeRedis_Del(t *testing.T) {
	f := &fakeRedis{store: map[string][]byte{"a": []byte("1"), "b": []byte("2")}}
	f.Del(context.Background(), "a", "b")
	if len(f.store) != 0 {
		t.Error("Del did not remove keys")
	}
}
