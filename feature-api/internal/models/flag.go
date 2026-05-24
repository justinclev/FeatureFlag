package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// RuleMatchStrategy defines how multiple rules are evaluated.
type RuleMatchStrategy string

const (
	// RuleMatchStrategyAny means the first matching rule determines the result (OR logic).
	RuleMatchStrategyAny RuleMatchStrategy = "any"
	// RuleMatchStrategyAll means all rules must match to return a value (AND logic).
	RuleMatchStrategyAll RuleMatchStrategy = "all"
)

// Flag is the domain model stored in MongoDB.
type Flag struct {
	ID                bson.ObjectID     `bson:"_id,omitempty"     json:"id"`
	Name               string            `bson:"name"              json:"name"`
	Description        string            `bson:"description"       json:"description"`
	Key                string            `bson:"key"               json:"key"`
	DefaultValue       bool              `bson:"defaultValue"      json:"defaultValue"`
	Rules              []Rule            `bson:"rules"             json:"rules"`
	RuleMatchStrategy  RuleMatchStrategy `bson:"ruleMatchStrategy" json:"ruleMatchStrategy"`
	CreatedAt          time.Time         `bson:"createdAt"         json:"createdAt"`
	CreatedBy          string            `bson:"createdBy"         json:"createdBy"`
	UpdatedAt          time.Time         `bson:"updatedAt"         json:"updatedAt"`
	UpdatedBy          string            `bson:"updatedBy"         json:"updatedBy"`
	Enabled            bool              `bson:"enabled"           json:"enabled"`
}

// Clone returns a true deep copy of the Flag.
func (f *Flag) Clone() *Flag {
	if f == nil {
		return nil
	}
	newFlag := *f
	if f.Rules != nil {
		newFlag.Rules = make([]Rule, len(f.Rules))
		for i, r := range f.Rules {
			newFlag.Rules[i] = r.Clone()
		}
	}
	return &newFlag
}

// Clone returns a deep copy of the Rule.
func (r Rule) Clone() Rule {
	newRule := r
	newRule.Config = r.Config.Clone()
	return newRule
}

// Clone returns a deep copy of RuleConfig, copying all pointers to new memory.
func (c RuleConfig) Clone() RuleConfig {
	newCfg := c

	// Deep copy slices
	if c.Countries != nil {
		newCfg.Countries = make([]string, len(c.Countries))
		copy(newCfg.Countries, c.Countries)
	}
	if c.States != nil {
		newCfg.States = make([]string, len(c.States))
		copy(newCfg.States, c.States)
	}
	if c.Cities != nil {
		newCfg.Cities = make([]string, len(c.Cities))
		copy(newCfg.Cities, c.Cities)
	}
	if c.ZipCodes != nil {
		newCfg.ZipCodes = make([]string, len(c.ZipCodes))
		copy(newCfg.ZipCodes, c.ZipCodes)
	}
	if c.UserIDs != nil {
		newCfg.UserIDs = make([]string, len(c.UserIDs))
		copy(newCfg.UserIDs, c.UserIDs)
	}

	// Deep copy pointers
	if c.Percentage != nil {
		v := *c.Percentage
		newCfg.Percentage = &v
	}
	if c.EnableAt != nil {
		v := *c.EnableAt
		newCfg.EnableAt = &v
	}
	if c.DisableAt != nil {
		v := *c.DisableAt
		newCfg.DisableAt = &v
	}
	if c.StartAt != nil {
		v := *c.StartAt
		newCfg.StartAt = &v
	}
	if c.EndAt != nil {
		v := *c.EndAt
		newCfg.EndAt = &v
	}
	if c.StartPercent != nil {
		v := *c.StartPercent
		newCfg.StartPercent = &v
	}
	if c.EndPercent != nil {
		v := *c.EndPercent
		newCfg.EndPercent = &v
	}

	return newCfg
}

// CreateFlagRequest defines the schema for creating a new feature flag.
type CreateFlagRequest struct {
	Key                string            `json:"key"`
	Name               string            `json:"name"`
	Enabled            bool              `json:"enabled"`
	Description        string            `json:"description"`
	DefaultValue       bool              `json:"defaultValue"`
	Rules              []Rule            `json:"rules"`
	RuleMatchStrategy  RuleMatchStrategy `json:"ruleMatchStrategy"`
	CreatedBy          string            `json:"createdBy"`
}

// UpdateFlagRequest defines the schema for updating an existing feature flag.
type UpdateFlagRequest struct {
	Key                *string            `json:"key,omitempty"`
	Name               *string            `json:"name,omitempty"`
	Enabled            *bool              `json:"enabled,omitempty"`
	Description        *string            `json:"description,omitempty"`
	DefaultValue       *bool              `json:"defaultValue,omitempty"`
	Rules              *[]Rule            `json:"rules,omitempty"`
	RuleMatchStrategy  *RuleMatchStrategy `json:"ruleMatchStrategy,omitempty"`
	UpdatedBy          string             `json:"updatedBy"`
}
