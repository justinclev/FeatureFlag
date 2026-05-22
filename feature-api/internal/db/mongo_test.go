package db

import (
	"os"
	"testing"

	"github.com/featureflags/feature-api/internal/config"
)

func TestConnectDisconnect(t *testing.T) {
	os.Setenv("API_KEY", "test")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB_NAME", "feature_flags_test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	client, db, err := Connect(cfg)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if db == nil {
		t.Error("expected db, got nil")
	}
	Disconnect(client)
}
