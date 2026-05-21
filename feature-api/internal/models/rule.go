package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

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
