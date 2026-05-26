package db

import (
	"testing"
	"github.com/featureflags/feature-api/internal/config"
)

func TestConnect_Error(t *testing.T) {
    cfg := &config.Config{MongoURI: "mongodb://localhost:1"}
    _, _, err := Connect(cfg)
    if err == nil {
        t.Error("expected ping error")
    }
}

func TestDisconnect_Nil(t *testing.T) {
    Disconnect(nil)
}

func TestConnect_InvalidURI(t *testing.T) {
    cfg := &config.Config{MongoURI: "mongodb://[::1]:invalid"}
    _, _, err := Connect(cfg)
    if err == nil {
        t.Error("expected error for invalid URI")
    }
}
