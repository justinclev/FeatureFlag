package repository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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
    // Optional targeted errors
    findOneErr error
    updateErr  error
    deleteErr  error
    pingErr    error
}

func (m *mockMongoCol) Find(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error) {
	if m.err != nil {
		return nil, m.err
	}
	return mongo.NewCursorFromDocuments(m.toDocuments(), nil, nil)
}

func (m *mockMongoCol) FindOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOneOptions]) *mongo.SingleResult {
	if m.findOneErr != nil {
		return mongo.NewSingleResultFromDocument(struct{}{}, m.findOneErr, nil)
	}
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
	if m.updateErr != nil {
		return mongo.NewSingleResultFromDocument(struct{}{}, m.updateErr, nil)
	}
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
	if m.deleteErr != nil {
		return nil, m.deleteErr
	}
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
    // Return a real database object with a fake client to test ping
    client, _ := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	return client.Database("test")
}

func (m *mockMongoCol) CountDocuments(ctx context.Context, filter interface{}, opts ...options.Lister[options.CountOptions]) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	f := filter.(bson.M)
	key, ok := f["key"].(string)
    if !ok {
        return 0, nil
    }
	var count int64
	for _, flag := range m.flags {
		if flag.Key == key {
			count++
		}
	}
	return count, nil
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
    setErr error
    getErr error
    delErr error
}

func (f *fakeRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	if f.getErr != nil {
		cmd.SetErr(f.getErr)
		return cmd
	}
	if f.err != nil {
		cmd.SetErr(f.err)
		return cmd
	}
	if val, ok := f.store[key]; ok {
		cmd.SetVal(string(val))
	} else {
		cmd.SetErr(redis.Nil)
	}
	return cmd
}

func (f *fakeRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if f.setErr != nil {
		cmd.SetErr(f.setErr)
		return cmd
	}
	if f.err != nil {
		cmd.SetErr(f.err)
		return cmd
	}
	f.store[key] = value.([]byte)
	return cmd
}

func (f *fakeRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
    if f.delErr != nil {
        cmd.SetErr(f.delErr)
        return cmd
    }
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

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestMongoRedisRepository_List(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{flags: []models.Flag{{ID: bson.NewObjectID(), Name: "flag1"}}}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, "test:")

	flags, err := repo.List(ctx, 10, 5)
	if err != nil || len(flags) != 1 || flags[0].Name != "flag1" {
		t.Errorf("expected 1 flag, got %v, %v", flags, err)
	}
}

func TestMongoRedisRepository_List_Empty(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{flags: nil}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, "test:")

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
	repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, "test:")

	flag, err := repo.GetByID(ctx, id.Hex())
	if err != nil || flag.Name != "flag1" {
		t.Errorf("expected flag1, got %v, %v", flag, err)
	}
}

func TestMongoRedisRepository_GetByKey(t *testing.T) {
	ctx := context.Background()
	key := "test-key"
	prefix := "test:"
	col := &mockMongoCol{flags: []models.Flag{{ID: bson.NewObjectID(), Key: key, Name: "flag1"}}}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, prefix)

	// Test cache miss
	flag, err := repo.GetByKey(ctx, key)
	if err != nil || flag.Name != "flag1" {
		t.Errorf("expected flag1, got %v, %v", flag, err)
	}

	// Test cache hit
	if _, ok := fakeRdb.store[prefix+key]; !ok {
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
	repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, "test:")

	req := models.CreateFlagRequest{Name: "new-flag", Key: "key1"}
	flag, err := repo.Create(ctx, req)
	if err != nil || flag.Name != "new-flag" {
		t.Errorf("expected new-flag, got %v, %v", flag, err)
	}
	if len(col.flags) != 1 {
		t.Errorf("expected 1 flag in DB, got %d", len(col.flags))
	}
}

func TestMongoRedisRepository_Create_Conflict(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{flags: []models.Flag{{Key: "exists"}}}
	fakeRdb := &fakeRedis{store: make(map[string][]byte)}
	repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, "test:")

	req := models.CreateFlagRequest{Key: "exists"}
	_, err := repo.Create(ctx, req)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestMongoRedisRepository_Update(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	key := "old-key"
	prefix := "test:"
	col := &mockMongoCol{flags: []models.Flag{{ID: id, Key: key, Name: "old-name"}}}
	fakeRdb := &fakeRedis{store: map[string][]byte{prefix+key: []byte(`{"name":"old-name","key":"old-key"}`)}}
	repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, prefix)

	name := "new-name"
	flag, err := repo.Update(ctx, id.Hex(), models.UpdateFlagRequest{Name: &name})
	if err != nil || flag.Name != "new-name" {
		t.Errorf("expected new-name, got %v, %v", flag, err)
	}

	// Cache should be invalidated
	if _, ok := fakeRdb.store[prefix+key]; ok {
		t.Error("expected cache invalidation after update")
	}
}

func TestMongoRedisRepository_Delete(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	key := "to-delete"
	prefix := "test:"
	col := &mockMongoCol{flags: []models.Flag{{ID: id, Key: key, Name: "to-delete"}}}
	fakeRdb := &fakeRedis{store: map[string][]byte{prefix+key: []byte(`{"name":"to-delete","key":"to-delete"}`)}}
	repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, prefix)

	err := repo.Delete(ctx, id.Hex())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Cache should be invalidated
	if _, ok := fakeRdb.store[prefix+key]; ok {
		t.Error("expected cache invalidation after delete")
	}
	if len(col.flags) != 0 {
		t.Errorf("expected 0 flags, got %d", len(col.flags))
	}
}

func TestMongoRedisRepository_GetByID_Errors(t *testing.T) {
	ctx := context.Background()
	repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{store: make(map[string][]byte)}, testLogger, time.Second, "test:")

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
	repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{}, testLogger, time.Second, "test:")

	// Invalid ID
	_, err := repo.Update(ctx, "invalid-hex", models.UpdateFlagRequest{})
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}

	// No fields
	id := bson.NewObjectID()
	col := &mockMongoCol{flags: []models.Flag{{ID: id}}}
	repoValid := NewMongoRedisRepository(col, &fakeRedis{}, testLogger, time.Second, "test:")
	_, err = repoValid.Update(ctx, id.Hex(), models.UpdateFlagRequest{})
	if !errors.Is(err, ErrNoFields) {
		t.Errorf("expected ErrNoFields, got %v", err)
	}
}

func TestMongoRedisRepository_Delete_Errors(t *testing.T) {
	ctx := context.Background()
	repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{}, testLogger, time.Second, "test:")

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

func TestMongoRedisRepository_List_ErrorPath(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{err: errors.New("db error")}
	repo := NewMongoRedisRepository(col, &fakeRedis{}, testLogger, time.Second, "test:")
	_, err := repo.List(ctx, 10, 0)
	if err == nil || !strings.Contains(err.Error(), "find flags") {
		t.Errorf("expected error, got %v", err)
	}
}

func TestMongoRedisRepository_GetByKey_UnmarshalFail(t *testing.T) {
	ctx := context.Background()
	prefix := "test:"
	fakeRdb := &fakeRedis{store: map[string][]byte{prefix+"fail": []byte("invalid-json")}}
    col := &mockMongoCol{flags: []models.Flag{{ID: bson.NewObjectID(), Key: "fail", Name: "ok"}}}
	repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, prefix)

	flag, err := repo.GetByKey(ctx, "fail")
	if err != nil || flag.Name != "ok" {
		t.Errorf("expected fallback to DB after unmarshal fail, got %v", err)
	}
}

func TestMongoRedisRepository_GetByKey_NegativeCache(t *testing.T) {
	ctx := context.Background()
	prefix := "test:"
    // Redis has the negative cache marker
	fakeRdb := &fakeRedis{store: map[string][]byte{prefix+"missing": []byte("__404__")}}
	repo := NewMongoRedisRepository(&mockMongoCol{}, fakeRdb, testLogger, time.Second, prefix)

	_, err := repo.GetByKey(ctx, "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound from negative cache, got %v", err)
	}
}

func TestMongoRedisRepository_Ready_RedisError(t *testing.T) {
	ctx := context.Background()
	fakeRdb := &fakeRedis{err: errors.New("redis fail")}
	repo := NewMongoRedisRepository(&mockMongoCol{}, fakeRdb, testLogger, time.Second, "test:")
	err := repo.Ready(ctx)
	if err == nil || !strings.Contains(err.Error(), "redis not ready") {
		t.Errorf("expected redis error, got %v", err)
	}
}

func TestMongoRedisRepository_GetByID_DBError(t *testing.T) {
	ctx := context.Background()
	col := &mockMongoCol{err: errors.New("db error")}
	repo := NewMongoRedisRepository(col, &fakeRedis{}, testLogger, time.Second, "test:")
	_, err := repo.GetByID(ctx, bson.NewObjectID().Hex())
	if err == nil || !strings.Contains(err.Error(), "find flag") {
		t.Errorf("expected error, got %v", err)
	}
}

func TestMongoRedisRepository_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	col := &mockMongoCol{} // empty
	repo := NewMongoRedisRepository(col, &fakeRedis{}, testLogger, time.Second, "test:")
	name := "new"
	_, err := repo.Update(ctx, id.Hex(), models.UpdateFlagRequest{Name: &name})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMongoRedisRepository_Update_DBError(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	col := &mockMongoCol{flags: []models.Flag{{ID: id}}, updateErr: errors.New("fail")}
	repo := NewMongoRedisRepository(col, &fakeRedis{}, testLogger, time.Second, "test:")
	name := "new"
	_, err := repo.Update(ctx, id.Hex(), models.UpdateFlagRequest{Name: &name})
	if err == nil || !strings.Contains(err.Error(), "update flag") {
		t.Errorf("expected update error, got %v", err)
	}
}

func TestMongoRedisRepository_Delete_DBError(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	col := &mockMongoCol{flags: []models.Flag{{ID: id}}, deleteErr: errors.New("fail")}
	repo := NewMongoRedisRepository(col, &fakeRedis{}, testLogger, time.Second, "test:")
	err := repo.Delete(ctx, id.Hex())
	if err == nil || !strings.Contains(err.Error(), "delete flag") {
		t.Errorf("expected delete error, got %v", err)
	}
}

func TestShardedL1Cache_Expired(t *testing.T) {
    cache := newShardedL1Cache()
    key := "key"
    cache.Set(key, &models.Flag{Name: "test"}, -1*time.Second) // already expired
    _, ok := cache.Get(key)
    if ok {
        t.Error("expected expired item to be missing")
    }
}

func TestShardedL1Cache_Remove(t *testing.T) {
    cache := newShardedL1Cache()
    key := "key"
    cache.Set(key, &models.Flag{Name: "test"}, time.Hour)
    cache.Remove(key)
    _, ok := cache.Get(key)
    if ok {
        t.Error("expected removed item to be missing")
    }
}

func TestShardedL1Cache_Cleanup(t *testing.T) {
    cache := newShardedL1Cache()
    cache.Set("expired", &models.Flag{Name: "old"}, -1*time.Second)
    cache.Set("valid", &models.Flag{Name: "new"}, time.Hour)
    
    cache.Cleanup()
    
    if _, ok := cache.Get("expired"); ok {
        t.Error("expected expired item to be removed by cleanup")
    }
    if _, ok := cache.Get("valid"); !ok {
        t.Error("expected valid item to stay after cleanup")
    }
}

func TestMongoRedisRepository_GetByKey_L1Hit(t *testing.T) {
    ctx := context.Background()
    repo := NewMongoRedisRepository(&mockMongoCol{}, &fakeRedis{}, testLogger, time.Second, "test:")
    flag := &models.Flag{Key: "hit"}
    repo.l1.Set("hit", flag, time.Hour)
    
    res, err := repo.GetByKey(ctx, "hit")
    if err != nil || res.Key != "hit" {
        t.Errorf("expected L1 hit, got %v, %v", res, err)
    }
}

func TestMongoRedisRepository_GetByKey_DBError(t *testing.T) {
    ctx := context.Background()
    col := &mockMongoCol{err: errors.New("db fail")}
    repo := NewMongoRedisRepository(col, &fakeRedis{}, testLogger, time.Second, "test:")
    
    _, err := repo.GetByKey(ctx, "key")
    if err == nil || !strings.Contains(err.Error(), "find flag") {
        t.Errorf("expected find error, got %v", err)
    }
}

func TestMongoRedisRepository_GetByKey_L2Hit(t *testing.T) {
    ctx := context.Background()
    prefix := "test:"
    flag := models.Flag{Key: "l2-hit", Name: "L2"}
    payload, _ := json.Marshal(flag)
    fakeRdb := &fakeRedis{store: map[string][]byte{prefix+"l2-hit": payload}}
    repo := NewMongoRedisRepository(&mockMongoCol{}, fakeRdb, testLogger, time.Second, prefix)
    
    res, err := repo.GetByKey(ctx, "l2-hit")
    if err != nil || res.Name != "L2" {
        t.Errorf("expected L2 hit, got %v, %v", res, err)
    }
    // Check if promoted to L1
    if _, ok := repo.l1.Get("l2-hit"); !ok {
        t.Error("expected promotion to L1")
    }
}

func TestMongoRedisRepository_Update_MultipleFields(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	key := "old"
	col := &mockMongoCol{flags: []models.Flag{{ID: id, Key: key}}}
	repo := NewMongoRedisRepository(col, &fakeRedis{store: make(map[string][]byte)}, testLogger, time.Second, "test:")

	name := "new"
    enabled := true
	_, err := repo.Update(ctx, id.Hex(), models.UpdateFlagRequest{Name: &name, Enabled: &enabled})
	if err != nil {
		t.Fatal(err)
	}
    if col.flags[0].Name != "new" || !col.flags[0].Enabled {
        t.Error("fields not updated")
    }
}

func TestMongoRedisRepository_RedisFailures(t *testing.T) {
    ctx := context.Background()
    prefix := "test:"
    col := &mockMongoCol{flags: []models.Flag{{Key: "key", Name: "ok"}}}
    fakeRdb := &fakeRedis{getErr: errors.New("redis fail"), setErr: errors.New("redis fail")}
    repo := NewMongoRedisRepository(col, fakeRdb, testLogger, time.Second, prefix)

    // GetByKey should still work via DB fallback
    flag, err := repo.GetByKey(ctx, "key")
    if err != nil || flag.Name != "ok" {
        t.Errorf("expected fallback to DB, got %v", err)
    }

    // Invalidate with Redis error should not panic
    repo.invalidate("key")
}
