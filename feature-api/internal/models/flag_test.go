package models

import (
	"testing"
	"time"
)

func TestFlagStruct(t *testing.T) {
	f := Flag{
		Name: "test",
		Key: "test-key",
		Enabled: true,
		DefaultValue: false,
		CreatedAt: time.Now(),
		CreatedBy: "user",
		UpdatedAt: time.Now(),
		UpdatedBy: "user",
	}
	if f.Name != "test" || f.Key != "test-key" || !f.Enabled {
		t.Error("flag struct fields not set correctly")
	}
}
