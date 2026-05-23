package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"testing"
	"time"

	"github.com/featureflags/feature-api/internal/models"
	"github.com/redis/go-redis/v9"
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

type mockMongoCol struct {
	flags []models.Flag
	err   error
}

func (m *mockMongoCol) Find(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error) {
	if m.err != nil {
		return nil, m.err
	}
	return mongo.NewCursorFromDocuments(m.toDocuments(), nil, nil)
}

func (m *mockMongoCol) FindOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOneOptions]) *mongo.SingleResult {
	if m.err != nil {
		return mongo.NewSingleResultFromDocument(bson.D{}, m.err, nil)
	}
	f, ok := filter.(bson.M)
	if !ok {
		return mongo.NewSingleResultFromDocument(bson.D{}, errors.New("invalid filter type"), nil)
	}
	id := f["_id"].(bson.ObjectID)
	for _, flag := range m.flags {
		if flag.ID == id {
			return mongo.NewSingleResultFromDocument(flag, nil, nil)
		}
	}
	return mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil)
}

func (m *mockMongoCol) InsertOne(ctx context.Context, document interface{}, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	flag := document.(models.Flag)
	m.flags = append(m.flags, flag)
	return &mongo.InsertOneResult{InsertedID: flag.ID}, nil
}

func (m *mockMongoCol) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult {
	if m.err != nil {
		return mongo.NewSingleResultFromDocument(bson.D{}, m.err, nil)
	}
	f := filter.(bson.M)
	id := f["_id"].(bson.ObjectID)
	for i, flag := range m.flags {
		if flag.ID == id {
			return mongo.NewSingleResultFromDocument(m.flags[i], nil, nil)
		}
	}
	return mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil)
}

func (m *mockMongoCol) DeleteOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	f := filter.(bson.M)
	id := f["_id"].(bson.ObjectID)
	for i, flag := range m.flags {
		if flag.ID == id {
			m.flags = append(m.flags[:i], m.flags[i+1:]...)
			return &mongo.DeleteResult{DeletedCount: 1}, nil
		}
	}
	return &mongo.DeleteResult{DeletedCount: 0}, nil
}

func (m *mockMongoCol) toDocuments() []interface{} {
	docs := make([]interface{}, len(m.flags))
	for i, f := range m.flags {
		docs[i] = f
	}
	return docs
}

func TestMongoRedisRepository_List(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{flags: []models.Flag{{ID: bson.NewObjectID(), Name: "flag1"}}}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second)

	flags, err := repo.List(ctx, 10, 5)
	if err != nil || len(flags) != 1 || flags[0].Name != "flag1" {
		t.Errorf("expected 1 flag, got %v, %v", flags, err)
	}
}

func TestMongoRedisRepository_List_Empty(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{flags: nil}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second)

	flags, err := repo.List(ctx, 0, 0)
	if err != nil || len(flags) != 0 {
		t.Errorf("expected empty flags, got %v, %v", flags, err)
	}
}

func TestMongoRedisRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	col := &mockMongoCol{flags: []models.Flag{{ID: id, Name: "flag1"}}}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second)

	// Test cache miss
	flag, err := repo.GetByID(ctx, id.Hex())
	if err != nil || flag.Name != "flag1" {
		t.Errorf("expected flag1, got %v, %v", flag, err)
	}

	// Test cache hit
	flag, err = repo.GetByID(ctx, id.Hex())
	if err != nil || flag.Name != "flag1" {
		t.Errorf("expected flag1 from cache, got %v, %v", flag, err)
	}
}

func TestMongoRedisRepository_Create(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second)

	flag, err := repo.Create(ctx, models.CreateFlagRequest{
		Name: "new-flag", 
		Key: "new-key",
		Rules: []models.Rule{{ID: bson.NewObjectID(), Type: models.RuleTypePercentage}},
	})
	if err != nil || flag.Name != "new-flag" {
		t.Errorf("failed to create: %v", err)
	}
	if len(col.flags) != 1 {
		t.Errorf("expected 1 flag in col, got %d", len(col.flags))
	}
}

func TestMongoRedisRepository_Update_Full(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	col := &mockMongoCol{flags: []models.Flag{{ID: id, Name: "old-name"}}}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second)

	name := "new-name"
	key := "new-key"
	enabled := true
	desc := "desc"
	defVal := true
	rules := []models.Rule{}
	req := models.UpdateFlagRequest{
		Name: &name,
		Key: &key,
		Enabled: &enabled,
		Description: &desc,
		DefaultValue: &defVal,
		Rules: &rules,
		UpdatedBy: "user",
	}
	flag, err := repo.Update(ctx, id.Hex(), req)
	if err != nil {
		t.Errorf("update failed: %v", err)
	}
	if flag == nil {
		t.Error("expected flag, got nil")
	}
}

func TestMongoRedisRepository_Delete(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	col := &mockMongoCol{flags: []models.Flag{{ID: id, Name: "to-delete"}}}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second)

	err := repo.Delete(ctx, id.Hex())
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}
	if len(col.flags) != 0 {
		t.Errorf("expected 0 flags, got %d", len(col.flags))
	}
}

func TestMongoRedisRepository_GetByID_Errors(t *testing.T) {
	ctx := context.Background()
	repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{store: make(map[string][]byte)}, time.Second)

	// Invalid ID
	_, err := repo.GetByID(ctx, "invalid-hex")
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}

	// Not found
	_, err = repo.GetByID(ctx, bson.NewObjectID().Hex())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Mongo error
	repoErr := NewMongoRedisRepository(&mockMongoCol{err: errors.New("mongo error")}, &fakeRedis{}, time.Second)
	_, err = repoErr.GetByID(ctx, bson.NewObjectID().Hex())
	if err == nil || !strings.Contains(err.Error(), "find flag") {
		t.Errorf("expected find flag error, got %v", err)
	}
}

func TestMongoRedisRepository_Update_Errors(t *testing.T) {
	ctx := context.Background()
	repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{}, time.Second)

	// Invalid ID
	_, err := repo.Update(ctx, "invalid-hex", models.UpdateFlagRequest{})
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}

	// No fields
	id := bson.NewObjectID()
	_, err = repo.Update(ctx, id.Hex(), models.UpdateFlagRequest{})
	if !errors.Is(err, ErrNoFields) {
		t.Errorf("expected ErrNoFields, got %v", err)
	}

	// Not found
	name := "new"
	_, err = repo.Update(ctx, id.Hex(), models.UpdateFlagRequest{Name: &name})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMongoRedisRepository_Delete_Errors(t *testing.T) {
	ctx := context.Background()
	repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{}, time.Second)

	// Invalid ID
	err := repo.Delete(ctx, "invalid-hex")
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}

	// Not found
	err = repo.Delete(ctx, bson.NewObjectID().Hex())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMongoRedisRepository_List_Error(t *testing.T) {
	ctx := context.Background()
	repo := NewMongoRedisRepository(&mockMongoCol{err: errors.New("fail")}, &fakeRedis{}, time.Second)
	_, err := repo.List(ctx, 10, 0)
	if err == nil || !strings.Contains(err.Error(), "find flags") {
		t.Errorf("expected error, got %v", err)
	}
}
