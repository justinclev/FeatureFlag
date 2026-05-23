package main

import (
	"os"
	"testing"
)

func TestRun_InvalidConfig(t *testing.T) {
	// Set an invalid config that will make run() fail early
	os.Setenv("API_KEY", "")
	err := run()
	if err == nil {
		t.Error("expected error due to missing API_KEY")
	}
}

func TestRun_InvalidMongoURI(t *testing.T) {
	os.Setenv("API_KEY", "test")
	os.Setenv("MONGO_URI", "invalid")
	err := run()
	if err == nil {
		t.Error("expected error for invalid mongo uri")
	}
}

func TestRun_InvalidRedisAddr(t *testing.T) {
    os.Setenv("API_KEY", "test")
    os.Setenv("MONGO_URI", "mongodb://localhost:27017")
    os.Setenv("REDIS_ADDR", "") // will fail Connect validation
    err := run()
    if err == nil {
        t.Error("expected error for empty redis addr")
    }
}
