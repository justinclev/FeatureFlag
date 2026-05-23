package db

import (
	"testing"
	"github.com/featureflags/feature-api/internal/config"
)

func TestDisconnect_Nil(t *testing.T) {
	Disconnect(nil)
}

func TestConnect_Error(t *testing.T) {
	cfg := &config.Config{MongoURI: "invalid-uri"}
	_, _, err := Connect(cfg)
	if err == nil {
		t.Error("expected error for invalid uri")
	}
}
