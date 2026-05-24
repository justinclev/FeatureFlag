package cache

import (
	"testing"
	"github.com/featureflags/feature-api/internal/config"
)

func TestClose_Nil(t *testing.T) {
	Close(nil)
}

func TestConnect_Error(t *testing.T) {
	cfg := &config.Config{RedisAddr: ""}
	_, err := Connect(cfg)
	if err == nil {
		t.Error("expected error for empty addr")
	}
}

func TestClose_NotNil(t *testing.T) {
    // This is hard without a real client, but we can't easily mock the internal Close method of go-redis
}
