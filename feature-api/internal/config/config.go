package config
package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration sourced from environment variables.
type Config struct {
	MongoURI    string
	MongoDBName string
	RedisAddr   string
	RedisPass   string
	Port        string
}

// Load reads environment variables, applies defaults, and validates required values.
func Load() (*Config, error) {
	cfg := &Config{
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB_NAME", "feature_flags"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASSWORD", ""),
		Port:        getEnv("PORT", "8080"),
	}

	if cfg.MongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI must not be empty")
	}
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR must not be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
