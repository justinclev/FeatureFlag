package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Rule defines a specific condition and its resulting value for flag evaluation.
type Rule struct {
	ID          bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	Description string         `bson:"description"   json:"description"`
	Type        RuleType       `bson:"type"          json:"type"`
	Config      map[string]any `bson:"config"        json:"config"`
	Value       bool           `bson:"value"         json:"value"`
}

// RuleType defines the strategy used for rule matching.
type RuleType string

const (
	RuleTypePercentage RuleType = "percentage"
	RuleTypeGeography  RuleType = "geography"
	RuleTypeSchedule   RuleType = "schedule"
	RuleTypeGradual    RuleType = "gradual"
	RuleTypeUserList   RuleType = "user_list"
	RuleTypeAttribute  RuleType = "attribute"
)

// Clone returns a deep copy of the Rule.
func (r Rule) Clone() Rule {
	newRule := r
	if r.Config != nil {
		newRule.Config = make(map[string]any)
		for k, v := range r.Config {
			newRule.Config[k] = v // Shallow copy of values is fine for primitives/strings
		}
	}
	return newRule
}
