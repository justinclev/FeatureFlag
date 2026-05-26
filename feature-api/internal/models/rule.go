package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"reflect"
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
	// RuleTypePercentage matches users based on a deterministic hash of their ID.
	RuleTypePercentage RuleType = "percentage"
	// RuleTypeGeography matches users based on location fields.
	RuleTypeGeography RuleType = "geography"
	// RuleTypeSchedule matches based on server-side time windows.
	RuleTypeSchedule RuleType = "schedule"
	// RuleTypeGradual matches based on a percentage that increases over time.
	RuleTypeGradual RuleType = "gradual"
	// RuleTypeUserList matches a specific list of user IDs.
	RuleTypeUserList RuleType = "user_list"
	// RuleTypeAttribute matches custom user attributes.
	RuleTypeAttribute RuleType = "attribute"
)

// Clone returns a deep copy of the Rule to prevent concurrent mutation of cached slices.
func (r Rule) Clone() Rule {
	newRule := r
	if r.Config != nil {
		newRule.Config = make(map[string]any, len(r.Config))
		for k, v := range r.Config {
			newRule.Config[k] = deepCopyValue(v)
		}
	}
	return newRule
}

func deepCopyValue(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsNil() {
			return nil
		}
		newSlice := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Cap())
		reflect.Copy(newSlice, rv)
		return newSlice.Interface()
	case reflect.Map:
		if rv.IsNil() {
			return nil
		}
		newMap := reflect.MakeMap(rv.Type())
		for _, key := range rv.MapKeys() {
			newMap.SetMapIndex(key, reflect.ValueOf(deepCopyValue(rv.MapIndex(key).Interface())))
		}
		return newMap.Interface()
	default:
		// Primitives and strings are safe to shallow copy
		return v
	}
}
