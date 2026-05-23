package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Rule defines a specific condition and its resulting value for flag evaluation.
type Rule struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Description string        `bson:"description"   json:"description"`
	Type        RuleType      `bson:"type"          json:"type"`
	Config      RuleConfig    `bson:"config"        json:"config"`
	Value       bool          `bson:"value"         json:"value"`
}

// RuleType defines the strategy used for rule matching.
type RuleType string

const (
	// RuleTypePercentage buckets users based on a static percentage.
	RuleTypePercentage RuleType = "percentage"
	// RuleTypeGeography matches users based on their location.
	RuleTypeGeography RuleType = "geography"
	// RuleTypeSchedule matches rules based on time windows.
	RuleTypeSchedule RuleType = "schedule"
	// RuleTypeGradual increases rollout percentage over time.
	RuleTypeGradual RuleType = "gradual"
	// RuleTypeUserList matches specific user IDs.
	RuleTypeUserList RuleType = "user_list"
	// RuleTypeAttribute matches custom user attributes.
	RuleTypeAttribute RuleType = "attribute"
)

// RuleConfig contains all potential parameters for different rule types.
type RuleConfig struct {
	Percentage     *float64   `bson:"percentage,omitempty"     json:"percentage,omitempty"`
	Countries      []string   `bson:"countries,omitempty"      json:"countries,omitempty"`
	States         []string   `bson:"states,omitempty"         json:"states,omitempty"`
	Cities         []string   `bson:"cities,omitempty"         json:"cities,omitempty"`
	ZipCodes       []string   `bson:"zipCodes,omitempty"       json:"zipCodes,omitempty"`
	EnableAt       *time.Time `bson:"enableAt,omitempty"       json:"enableAt,omitempty"`
	DisableAt      *time.Time `bson:"disableAt,omitempty"      json:"disableAt,omitempty"`
	StartAt        *time.Time `bson:"startAt,omitempty"        json:"startAt,omitempty"`
	EndAt          *time.Time `bson:"endAt,omitempty"          json:"endAt,omitempty"`
	StartPercent   *float64   `bson:"startPercent,omitempty"   json:"startPercent,omitempty"`
	EndPercent     *float64   `bson:"endPercent,omitempty"     json:"endPercent,omitempty"`
	UserIDs        []string   `bson:"userIds,omitempty"        json:"userIds,omitempty"`
	AttributeKey   string     `bson:"attributeKey,omitempty"   json:"attributeKey,omitempty"`
	AttributeOp    string     `bson:"attributeOp,omitempty"    json:"attributeOp,omitempty"`
	AttributeValue string     `bson:"attributeValue,omitempty" json:"attributeValue,omitempty"`
}
