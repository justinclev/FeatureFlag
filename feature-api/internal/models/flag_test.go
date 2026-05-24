package models

import (
	"testing"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestFlag_Clone(t *testing.T) {
	id := bson.NewObjectID()
	rules := []Rule{{Type: RuleTypePercentage, Value: true}}
	f := &Flag{
		ID:    id,
		Name:  "original",
		Rules: rules,
	}

	cloned := f.Clone()

	if cloned == f {
		t.Error("expected different memory address for clone")
	}
	if cloned.ID != f.ID || cloned.Name != f.Name {
		t.Error("cloned fields do not match original")
	}
	if len(cloned.Rules) != len(f.Rules) {
		t.Error("cloned rules length mismatch")
	}
	
	// Test deep copy of rules
	cloned.Rules[0].Value = false
	if f.Rules[0].Value == false {
		t.Error("original rules modified by clone change (shallow copy detected)")
	}

	// Test nil clone
	var nilFlag *Flag
	if nilFlag.Clone() != nil {
		t.Error("nil flag should clone to nil")
	}
}
