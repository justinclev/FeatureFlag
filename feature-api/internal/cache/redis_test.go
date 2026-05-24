package cache

import (
	"testing"

	"github.com/featureflags/feature-api/internal/config"
)

func TestConnect_Error(t *testing.T) {
	cfg := &config.Config{RedisAddr: ""}
	_, err := Connect(cfg)
	if err == nil {
		t.Error("expected error for empty address")
	}

	cfg.RedisAddr = "localhost:1" // likely to fail
	_, err = Connect(cfg)
	if err == nil {
		t.Error("expected ping error")
	}
}

func TestClose(t *testing.T) {
	Close(nil)
}
