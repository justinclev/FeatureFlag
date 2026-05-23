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
