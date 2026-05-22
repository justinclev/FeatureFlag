package cache

import (
	"os"
	"testing"

	"github.com/featureflags/feature-api/internal/config"
)

func TestConnectClose(t *testing.T) {
	os.Setenv("API_KEY", "test")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB_NAME", "test")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	client, err := Connect(cfg)
	if err != nil {
		t.Fatalf("redis connect: %v", err)
	}
	Close(client)
}
