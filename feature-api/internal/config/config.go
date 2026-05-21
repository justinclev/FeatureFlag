package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration sourced from environment variables.
type Config struct {
	MongoURI          string
	MongoDBName       string
	RedisAddr         string
	RedisPass         string
	Port              string
	CORSAllowedOrigin string
	CacheTTL          time.Duration
	APIKey            string
}

// Load reads environment variables, applies defaults, and validates required values.
func Load() (*Config, error) {
	cfg := &Config{
		MongoURI:          getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:       getEnv("MONGO_DB_NAME", "feature_flags"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:         getEnv("REDIS_PASSWORD", ""),
		Port:              getEnv("PORT", "8080"),
		CORSAllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:4200"),
		CacheTTL:          time.Duration(getEnvInt("CACHE_TTL_SECONDS", 30)) * time.Second,
		APIKey:            getEnv("API_KEY", ""),
	}

	if cfg.MongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI must not be empty")
	}
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR must not be empty")
	}
	if cfg.CacheTTL <= 0 {
		return nil, fmt.Errorf("CACHE_TTL_SECONDS must be a positive integer")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API_KEY must not be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
