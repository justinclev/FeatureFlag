package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/featureflags/feature-api/internal/models"
)

// mockMongoCol implements MongoCollection for testing.
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
		return mongo.NewSingleResultFromDocument(struct{}{}, m.err, nil)
	}
	f := filter.(bson.M)
	if id, ok := f["_id"].(bson.ObjectID); ok {
		for _, flag := range m.flags {
			if flag.ID == id {
				return mongo.NewSingleResultFromDocument(flag, nil, nil)
			}
		}
	}
	if key, ok := f["key"].(string); ok {
		for _, flag := range m.flags {
			if flag.Key == key {
				return mongo.NewSingleResultFromDocument(flag, nil, nil)
			}
		}
	}
	return mongo.NewSingleResultFromDocument(struct{}{}, mongo.ErrNoDocuments, nil)
}

func (m *mockMongoCol) InsertOne(ctx context.Context, document interface{}, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	f := document.(models.Flag)
	m.flags = append(m.flags, f)
	return &mongo.InsertOneResult{InsertedID: f.ID}, nil
}

func (m *mockMongoCol) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult {
	if m.err != nil {
		return mongo.NewSingleResultFromDocument(struct{}{}, m.err, nil)
	}
	f := filter.(bson.M)
	id := f["_id"].(bson.ObjectID)
	for i, flag := range m.flags {
		if flag.ID == id {
			u := update.(bson.M)["$set"].(bson.M)
			if n, ok := u["name"].(string); ok {
				m.flags[i].Name = n
			}
			if k, ok := u["key"].(string); ok {
				m.flags[i].Key = k
			}
			if e, ok := u["enabled"].(bool); ok {
				m.flags[i].Enabled = e
			}
			if d, ok := u["description"].(string); ok {
				m.flags[i].Description = d
			}
			if dv, ok := u["defaultValue"].(bool); ok {
				m.flags[i].DefaultValue = dv
			}
			if r, ok := u["rules"].([]models.Rule); ok {
				m.flags[i].Rules = r
			}
			if s, ok := u["ruleMatchStrategy"].(models.RuleMatchStrategy); ok {
				m.flags[i].RuleMatchStrategy = s
			}
			return mongo.NewSingleResultFromDocument(m.flags[i], nil, nil)
		}
	}
	return mongo.NewSingleResultFromDocument(struct{}{}, mongo.ErrNoDocuments, nil)
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

func (m *mockMongoCol) Database() *mongo.Database {
	return nil
}

func (m *mockMongoCol) toDocuments() []interface{} {
	docs := make([]interface{}, len(m.flags))
	for i, f := range m.flags {
		docs[i] = f
	}
	return docs
}

// fakeRedis implements RedisClient for testing.
type fakeRedis struct {
	store map[string][]byte
	err   error
}

func (f *fakeRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	if val, ok := f.store[key]; ok {
		cmd.SetVal(string(val))
	} else {
		cmd.SetErr(redis.Nil)
	}
	return cmd
}

func (f *fakeRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	f.store[key] = value.([]byte)
	return cmd
}

func (f *fakeRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	for _, k := range keys {
		delete(f.store, k)
	}
	return cmd
}

func (f *fakeRedis) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if f.err != nil {
		cmd.SetErr(f.err)
	} else {
		cmd.SetVal("PONG")
	}
	return cmd
}

func TestMongoRedisRepository_List(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{flags: []models.Flag{{ID: bson.NewObjectID(), Name: "flag1"}}}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second, "test:")

	flags, err := repo.List(ctx, 10, 5)
	if err != nil || len(flags) != 1 || flags[0].Name != "flag1" {
		t.Errorf("expected 1 flag, got %v, %v", flags, err)
	}
}

func TestMongoRedisRepository_List_Empty(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{flags: nil}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second, "test:")

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
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second, "test:")

	flag, err := repo.GetByID(ctx, id.Hex())
	if err != nil || flag.Name != "flag1" {
		t.Errorf("expected flag1, got %v, %v", flag, err)
	}
}

func TestMongoRedisRepository_GetByKey(t *testing.T) {
	ctx := context.Background()
	key := "test-key"
	col := &mockMongoCol{flags: []models.Flag{{ID: bson.NewObjectID(), Key: key, Name: "flag1"}}}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second, "test:")

	// Test cache miss
	flag, err := repo.GetByKey(ctx, key)
	if err != nil || flag.Name != "flag1" {
		t.Errorf("expected flag1, got %v, %v", flag, err)
	}

	// Test cache hit
	if _, ok := fakeRdb.store["flags:key:"+key]; !ok {
		t.Error("expected flag to be cached in Redis")
	}

	flagCached, err := repo.GetByKey(ctx, key)
	if err != nil || flagCached.Name != "flag1" {
		t.Errorf("expected flag1 from cache, got %v, %v", flagCached, err)
	}
}

func TestMongoRedisRepository_Create(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second, "test:")

	req := models.CreateFlagRequest{Name: "new-flag", Key: "key1"}
	flag, err := repo.Create(ctx, req)
	if err != nil || flag.Name != "new-flag" {
		t.Errorf("expected new-flag, got %v, %v", flag, err)
	}
	if len(col.flags) != 1 {
		t.Errorf("expected 1 flag in DB, got %d", len(col.flags))
	}
}

func TestMongoRedisRepository_Update(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	key := "old-key"
	col := &mockMongoCol{flags: []models.Flag{{ID: id, Key: key, Name: "old-name"}}}
	fakeRdb := &fakeRedis{store: map[string][]byte{"flags:key:"+key: []byte(`{"name":"old-name","key":"old-key"}`)}}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second, "test:")

	name := "new-name"
	flag, err := repo.Update(ctx, id.Hex(), models.UpdateFlagRequest{Name: &name})
	if err != nil || flag.Name != "new-name" {
		t.Errorf("expected new-name, got %v, %v", flag, err)
	}

	// Cache should be invalidated
	if _, ok := fakeRdb.store["flags:key:"+key]; ok {
		t.Error("expected cache invalidation after update")
	}
}

func TestMongoRedisRepository_Delete(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	key := "to-delete"
	col := &mockMongoCol{flags: []models.Flag{{ID: id, Key: key, Name: "to-delete"}}}
	fakeRdb := &fakeRedis{store: map[string][]byte{"flags:key:"+key: []byte(`{"name":"to-delete","key":"to-delete"}`)}}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second, "test:")

	err := repo.Delete(ctx, id.Hex())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Cache should be invalidated
	if _, ok := fakeRdb.store["flags:key:"+key]; ok {
		t.Error("expected cache invalidation after delete")
	}
	if len(col.flags) != 0 {
		t.Errorf("expected 0 flags, got %d", len(col.flags))
	}
}

func TestMongoRedisRepository_GetByID_Errors(t *testing.T) {
	ctx := context.Background()
	repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{store: make(map[string][]byte)}, time.Second, "test:")

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
}

func TestMongoRedisRepository_Update_Errors(t *testing.T) {
	ctx := context.Background()
	repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{}, time.Second, "test:")

	// Invalid ID
	_, err := repo.Update(ctx, "invalid-hex", models.UpdateFlagRequest{})
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}

	// No fields
	id := bson.NewObjectID()
	col := &mockMongoCol{flags: []models.Flag{{ID: id}}}
	repoValid := NewMongoRedisRepository(col, &fakeRedis{}, time.Second, "test:")
	_, err = repoValid.Update(ctx, id.Hex(), models.UpdateFlagRequest{})
	if !errors.Is(err, ErrNoFields) {
		t.Errorf("expected ErrNoFields, got %v", err)
	}
}

func TestMongoRedisRepository_Delete_Errors(t *testing.T) {
	ctx := context.Background()
	repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{}, time.Second, "test:")

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
	repo := NewMongoRedisRepository(&mockMongoCol{err: errors.New("fail")}, &fakeRedis{}, time.Second, "test:")
	_, err := repo.List(ctx, 10, 0)
	if err == nil || !strings.Contains(err.Error(), "find flags") {
		t.Errorf("expected error, got %v", err)
	}
}

func TestMongoRedisRepository_GetByKey_UnmarshalError(t *testing.T) {
	ctx := context.Background()
	key := "test-key"
	col := &mockMongoCol{flags: []models.Flag{{ID: bson.NewObjectID(), Key: key, Name: "flag1"}}}
	// Redis returns garbage JSON
	fakeRdb := &fakeRedis{store: map[string][]byte{"flags:key:"+key: []byte("invalid-json")}}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second, "test:")

	// Should fall back to DB
	flag, err := repo.GetByKey(ctx, key)
	if err != nil || flag.Name != "flag1" {
		t.Errorf("expected fallback to DB, got %v, %v", flag, err)
	}
}

func TestMongoRedisRepository_Create_Error(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{err: errors.New("insert fail")}
	repo := NewMongoRedisRepository(col, &fakeRedis{}, time.Second, "test:")
	_, err := repo.Create(ctx, models.CreateFlagRequest{Name: "fail", Key: "fail"})
	if err == nil || !strings.Contains(err.Error(), "insert flag") {
		t.Errorf("expected error, got %v", err)
	}
}

func TestMongoRedisRepository_Update_AllFields(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	key := "old"
	col := &mockMongoCol{flags: []models.Flag{{ID: id, Key: key, Name: "old"}}}
	repo := NewMongoRedisRepository(col, &fakeRedis{store: make(map[string][]byte)}, time.Second, "test:")

	name := "new-name"
	newKey := "new-key"
	enabled := true
	desc := "new-desc"
	defVal := true
	rules := []models.Rule{{Type: models.RuleTypePercentage}}
	strategy := models.RuleMatchStrategyAll

	req := models.UpdateFlagRequest{
		Name:              &name,
		Key:               &newKey,
		Enabled:           &enabled,
		Description:       &desc,
		DefaultValue:      &defVal,
		Rules:             &rules,
		RuleMatchStrategy: &strategy,
		UpdatedBy:         "tester",
	}

	flag, err := repo.Update(ctx, id.Hex(), req)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if flag.Name != name || flag.Key != newKey || flag.Enabled != enabled || flag.RuleMatchStrategy != strategy {
		t.Errorf("some fields were not updated correctly: %+v", flag)
	}
}

func TestMongoRedisRepository_Ready_RedisError(t *testing.T) {
	ctx := context.Background()
	fakeRdb := &fakeRedis{err: errors.New("redis fail")}
	repo := NewMongoRedisRepository(&mockMongoCol{}, fakeRdb, time.Second, "test:")
	err := repo.Ready(ctx)
	if err == nil || !strings.Contains(err.Error(), "redis not ready") {
		t.Errorf("expected redis error, got %v", err)
	}
}

func TestMongoRedisRepository_GetByKey_UnmarshalFail(t *testing.T) {
	ctx := context.Background()
	fakeRdb := &fakeRedis{store: map[string][]byte{"flags:key:fail": []byte("invalid-json")}}
    col := &mockMongoCol{flags: []models.Flag{{ID: bson.NewObjectID(), Key: "fail", Name: "ok"}}}
	repo := NewMongoRedisRepository(col, fakeRdb, time.Second, "test:")

	flag, err := repo.GetByKey(ctx, "fail")
	if err != nil || flag.Name != "ok" {
		t.Errorf("expected fallback to DB after unmarshal fail, got %v", err)
	}
}

func TestMongoRedisRepository_List_ErrorPath(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{err: errors.New("db error")}
	repo := NewMongoRedisRepository(col, &fakeRedis{}, time.Second, "test:")
	_, err := repo.List(ctx, 10, 0)
	if err == nil || !strings.Contains(err.Error(), "find flags") {
		t.Errorf("expected error, got %v", err)
	}
}
