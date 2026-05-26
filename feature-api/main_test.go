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

func TestRun_LogLevels(t *testing.T) {
	levels := []string{"debug", "warn", "error", "info", "invalid"}
	for _, l := range levels {
		os.Setenv("LOG_LEVEL", l)
		os.Setenv("API_KEY", "test")
		os.Setenv("MONGO_URI", "invalid")
		_ = run() // will still fail at mongo connect, but cover the switch
	}
}
