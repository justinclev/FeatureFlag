package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Flag is the domain model stored in MongoDB.
type Flag struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name"          json:"name"`
	Description  string        `bson:"description"   json:"description"`
	Key          string        `bson:"key"          json:"key"`
	DefaultValue bool          `bson:"defaultValue" json:"defaultValue"`
	Rules        []Rule        `bson:"rules"        json:"rules"`
	CreatedAt    time.Time     `bson:"createdAt"    json:"createdAt"`
	CreatedBy    string        `bson:"createdBy"    json:"createdBy"`
	UpdatedAt    time.Time     `bson:"updatedAt"    json:"updatedAt"`
	UpdatedBy    string        `bson:"updatedBy"    json:"updatedBy"`
	Enabled      bool          `bson:"enabled"      json:"enabled"`
}

type Rule struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Description string        `bson:"description"   json:"description"`
	Type        RuleType      `bson:"type"          json:"type"`
	Config      RuleConfig    `bson:"config"        json:"config"`
	Value       bool          `bson:"value"         json:"value"`
}

type RuleType string

const (
	RuleTypePercentage RuleType = "percentage"
	RuleTypeGeography  RuleType = "geography"
	RuleTypeSchedule   RuleType = "schedule"
	RuleTypeGradual    RuleType = "gradual"
	RuleTypeUserList   RuleType = "user_list"
	RuleTypeAttribute  RuleType = "attribute"
)

type RuleConfig struct {
	Percentage     *float64   `bson:"percentage,omitempty"   json:"percentage,omitempty"`
	Countries      []string   `bson:"countries,omitempty"    json:"countries,omitempty"`
	EnableAt       *time.Time `bson:"enableAt,omitempty"     json:"enableAt,omitempty"`
	DisableAt      *time.Time `bson:"disableAt,omitempty"    json:"disableAt,omitempty"`
	StartAt        *time.Time `bson:"startAt,omitempty"      json:"startAt,omitempty"`
	EndAt          *time.Time `bson:"endAt,omitempty"        json:"endAt,omitempty"`
	StartPercent   *float64   `bson:"startPercent,omitempty" json:"startPercent,omitempty"`
	EndPercent     *float64   `bson:"endPercent,omitempty"   json:"endPercent,omitempty"`
	UserIDs        []string   `bson:"userIds,omitempty"        json:"userIds,omitempty"`
	AttributeKey   string     `bson:"attributeKey,omitempty"   json:"attributeKey,omitempty"`
	AttributeOp    string     `bson:"attributeOp,omitempty"    json:"attributeOp,omitempty"`
	AttributeValue string     `bson:"attributeValue,omitempty" json:"attributeValue,omitempty"`
	States         []string   `bson:"states,omitempty"   json:"states,omitempty"`
	ZipCodes       []string   `bson:"zipCodes,omitempty" json:"zipCodes,omitempty"`
	Cities         []string   `bson:"cities,omitempty"   json:"cities,omitempty"`
}

// EvaluationContext contains user information for flag evaluation.
type EvaluationContext struct {
	UserID     string            `bson:"userId" json:"userId"`         // used for percentage hashing
	Country    string            `bson:"country" json:"country"`       // ISO 3166-1 alpha-2
	State      string            `bson:"state" json:"state"`           // US state code
	City       string            `bson:"city" json:"city"`             // city name
	ZipCode    string            `bson:"zipCode" json:"zipCode"`       // postal code
	Attributes map[string]string `bson:"attributes" json:"attributes"` // additional attributes for future extensibility
}

// EvaluationResult contains the result of flag evaluation.
type EvaluationResult struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}

type CreateFlagRequest struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	Description  string `json:"description"`
	DefaultValue bool   `json:"defaultValue"`
	Rules        []Rule `json:"rules"`
	CreatedBy    string `json:"createdBy"`
}

type UpdateFlagRequest struct {
	Key          *string `json:"key,omitempty"`
	Name         *string `json:"name,omitempty"`
	Enabled      *bool   `json:"enabled,omitempty"`
	Description  *string `json:"description,omitempty"`
	DefaultValue *bool   `json:"defaultValue,omitempty"`
	Rules        *[]Rule `json:"rules,omitempty"`
	UpdatedBy    string  `json:"updatedBy"`
}
