package models

import (
	"testing"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRuleClone(t *testing.T) {
	id := bson.NewObjectID()
	cfg := map[string]any{"foo": "bar"}
	r := Rule{
		ID:     id,
		Type:   RuleTypeAttribute,
		Config: cfg,
		Value:  true,
	}

	cloned := r.Clone()

	if cloned.ID != r.ID || cloned.Value != r.Value {
		t.Error("cloned fields mismatch")
	}
	
	// Test map isolation
	cloned.Config["foo"] = "baz"
	if r.Config["foo"] == "baz" {
		t.Error("original config modified by clone change")
	}
}
