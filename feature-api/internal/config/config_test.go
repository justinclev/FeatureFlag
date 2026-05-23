package config

import (
	"os"
	"testing"
	"time"
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
	if cfg.CacheTTL != 30*time.Second {
		t.Errorf("expected CacheTTL=30s, got %v", cfg.CacheTTL)
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("API_KEY", "test")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Unsetenv("MONGO_DB_NAME")
	os.Unsetenv("REDIS_ADDR")
	os.Unsetenv("CACHE_TTL_SECONDS")
	
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cfg.MongoDBName != "feature_flags" {
		t.Errorf("expected default DB name, got %q", cfg.MongoDBName)
	}
	if cfg.RedisAddr != "localhost:6379" {
		t.Errorf("expected default Redis addr, got %q", cfg.RedisAddr)
	}
	if cfg.CacheTTL != 30*time.Second {
		t.Errorf("expected default TTL 30s, got %v", cfg.CacheTTL)
	}
}

func TestLoad_MissingAPIKey(t *testing.T) {
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Unsetenv("API_KEY")
	_, err := Load()
	if err == nil {
		t.Error("expected error for missing API_KEY")
	}
}

func TestLoad_InvalidTTL(t *testing.T) {
	os.Setenv("API_KEY", "test")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("CACHE_TTL_SECONDS", "invalid")
	
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected success (defaulting on invalid), got %v", err)
	}
	if cfg.CacheTTL != 30*time.Second {
		t.Errorf("expected default TTL for invalid input, got %v", cfg.CacheTTL)
	}
}

func TestLoad_NegativeTTL(t *testing.T) {
	os.Setenv("API_KEY", "test")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("CACHE_TTL_SECONDS", "-10")
	
	_, err := Load()
	if err == nil {
		t.Error("expected error for negative TTL")
	}
}
