package models

import "testing"

func TestEvaluationContextStruct(t *testing.T) {
	ec := EvaluationContext{UserID: "u", Country: "US", Attributes: map[string]any{"foo": "bar"}}
	if ec.UserID != "u" || ec.Country != "US" || ec.Attributes["foo"] != "bar" {
		t.Error("evaluation context fields not set correctly")
	}
}

func TestEvaluationResultStruct(t *testing.T) {
	er := EvaluationResult{Enabled: true, Reason: "test"}
	if !er.Enabled || er.Reason != "test" {
		t.Error("evaluation result fields not set correctly")
	}
}
