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
	OffValue           bool              `bson:"offValue"          json:"offValue"`
	FallthroughValue   bool              `bson:"fallthroughValue"  json:"fallthroughValue"`
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

// CreateFlagRequest defines the schema for creating a new feature flag.
type CreateFlagRequest struct {
	Key                string            `json:"key"`
	Name               string            `json:"name"`
	Enabled            bool              `json:"enabled"`
	Description        string            `json:"description"`
	OffValue           bool              `json:"offValue"`
	FallthroughValue   bool              `json:"fallthroughValue"`
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
	OffValue           *bool              `json:"offValue,omitempty"`
	FallthroughValue   *bool              `json:"fallthroughValue,omitempty"`
	Rules              *[]Rule            `json:"rules,omitempty"`
	RuleMatchStrategy  *RuleMatchStrategy `json:"ruleMatchStrategy,omitempty"`
	UpdatedBy          string             `json:"updatedBy"`
}
