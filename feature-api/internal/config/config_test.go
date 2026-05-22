package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	os.Setenv("API_KEY", "test")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB_NAME", "feature_flags_test")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("CACHE_TTL_SECONDS", "30")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cfg.MongoURI == "" || cfg.RedisAddr == "" || cfg.APIKey == "" {
		t.Error("missing required config fields")
	}
}

func TestLoad_MissingAPIKey(t *testing.T) {
		os.Setenv("API_KEY", "test")
	os.Unsetenv("API_KEY")
	_, err := Load()
	if err == nil {
		t.Error("expected error for missing API_KEY")
	}
}
