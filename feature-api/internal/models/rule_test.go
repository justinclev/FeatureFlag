package models

import (
	"testing"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRuleClone(t *testing.T) {
	id := bson.NewObjectID()
	cfg := map[string]any{
		"foo": "bar",
		"list": []string{"a", "b"},
		"nested": map[string]any{
			"key": "val",
			"sublist": []int{1, 2},
		},
	}
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
		t.Error("original config modified by clone change (shallow map copy)")
	}

	// Test slice isolation (Deep copy check)
	clonedList := cloned.Config["list"].([]string)
	clonedList[0] = "Z"
	origList := r.Config["list"].([]string)
	if origList[0] == "Z" {
		t.Error("original slice modified by clone change (shallow slice copy)")
	}

	// Test nested map isolation
	clonedNested := cloned.Config["nested"].(map[string]any)
	clonedNested["key"] = "changed"
	origNested := r.Config["nested"].(map[string]any)
	if origNested["key"] == "changed" {
		t.Error("original nested map modified by clone change")
	}

	// Test nested slice isolation
	clonedSublist := clonedNested["sublist"].([]int)
	clonedSublist[0] = 99
	origSublist := origNested["sublist"].([]int)
	if origSublist[0] == 99 {
		t.Error("original nested slice modified by clone change")
	}
}

func TestRuleClone_NilConfig(t *testing.T) {
	r := Rule{ID: bson.NewObjectID(), Config: nil}
	cloned := r.Clone()
	if cloned.Config != nil {
		t.Error("expected nil config in clone")
	}
}

func TestDeepCopyValue_Nil(t *testing.T) {
    if deepCopyValue(nil) != nil {
        t.Error("expected nil for nil input")
    }
}
