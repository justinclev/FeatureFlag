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
	ID                 bson.ObjectID     `bson:"_id,omitempty"     json:"id"`
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
