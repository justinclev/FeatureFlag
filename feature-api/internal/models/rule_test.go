package models

import "testing"

func TestRuleTypeConstants(t *testing.T) {
	if RuleTypePercentage != "percentage" || RuleTypeGeography != "geography" {
		t.Error("rule type constants incorrect")
	}
}

func TestRuleConfigStruct(t *testing.T) {
	cfg := RuleConfig{AttributeKey: "foo", AttributeOp: "eq", AttributeValue: "bar"}
	if cfg.AttributeKey != "foo" || cfg.AttributeOp != "eq" || cfg.AttributeValue != "bar" {
		t.Error("rule config fields not set correctly")
	}
}
