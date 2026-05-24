package evaluator

import (
	"testing"
	"time"

	"github.com/featureflags/feature-api/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var eval = New()

func flagWith(rules []models.Rule, defaultValue bool, enabled bool) models.Flag {
	return models.Flag{
		ID:           bson.NewObjectID(),
		Key:          "test-flag",
		Name:         "Test Flag",
		Enabled:      enabled,
		DefaultValue: defaultValue,
		Rules:        rules,
	}
}

func ruleWith(t models.RuleType, cfg map[string]any, value bool) models.Rule {
	return models.Rule{
		ID:     bson.NewObjectID(),
		Type:   t,
		Config: cfg,
		Value:  value,
	}
}

// --- Flag-level ---

func TestEvaluate_FlagDisabled(t *testing.T) {
	flag := flagWith(nil, true, false)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"})
	if result.Enabled {
		t.Error("expected disabled")
	}
	if result.Reason != "flag disabled" {
		t.Errorf("unexpected reason: %q", result.Reason)
	}
}

func TestEvaluate_NoRules_DefaultTrue(t *testing.T) {
	flag := flagWith(nil, true, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{})
	if !result.Enabled {
		t.Error("expected enabled via default")
	}
	if result.Reason != "default value (no rules)" {
		t.Errorf("unexpected reason: %q", result.Reason)
	}
}

// --- Percentage ---

func TestEvaluate_Percentage_AlwaysMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypePercentage, map[string]any{"percentage": 100.0}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"})
	if !result.Enabled {
		t.Error("expected match with 100% rollout")
	}
}

func TestEvaluate_Percentage_NeverMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypePercentage, map[string]any{"percentage": 0.0}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"})
	if result.Enabled {
		t.Error("expected no match with 0% rollout")
	}
}

// --- UserList ---

func TestEvaluate_UserList_Match(t *testing.T) {
	rule := ruleWith(models.RuleTypeUserList, map[string]any{"userIds": []any{"user-1", "user-2"}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-2"})
	if !result.Enabled {
		t.Error("expected match for user in list")
	}
}

func TestEvaluate_UserList_ShardedMatch(t *testing.T) {
    // Large list (>20) to hit the map optimization
    uids := make([]any, 30)
    for i := 0; i < 30; i++ { uids[i] = "user-" + toString(i) }
    rule := ruleWith(models.RuleTypeUserList, map[string]any{"userIds": uids}, true)
    flag := flagWith([]models.Rule{rule}, false, true)
    if !eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-15"}).Enabled {
        t.Error("expected sharded user_list match")
    }
}

// --- Geography ---

func TestEvaluate_Geography_CountryMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, map[string]any{"countries": []any{"US"}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{Country: "us"})
	if !result.Enabled {
		t.Error("expected country match (EqualFold)")
	}
}

func TestEvaluate_Geography_FullMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, map[string]any{
        "countries": []any{"US"},
        "states": []any{"CA"},
        "cities": []any{"SF"},
        "zipCodes": []any{"94105"},
    }, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	ctx := models.EvaluationContext{Country: "US", State: "CA", City: "SF", ZipCode: "94105"}
	if !eval.Evaluate(&flag, ctx).Enabled {
		t.Error("expected full geography match")
	}
}

// --- Schedule ---

func TestEvaluate_Schedule_Valid(t *testing.T) {
	now := time.Now().UTC()
	rule := ruleWith(models.RuleTypeSchedule, map[string]any{
		"enableAt":  now.Add(-1 * time.Hour).Format(time.RFC3339),
		"disableAt": now.Add(1 * time.Hour).Format(time.RFC3339),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if !eval.Evaluate(&flag, models.EvaluationContext{}).Enabled {
		t.Error("expected match within schedule window")
	}
}

// --- Gradual ---

func TestEvaluate_Gradual_AlwaysMatch(t *testing.T) {
	now := time.Now().UTC()
	rule := ruleWith(models.RuleTypeGradual, map[string]any{
		"startAt":      now.Add(-1 * time.Hour).Format(time.RFC3339),
		"endAt":        now.Add(1 * time.Hour).Format(time.RFC3339),
		"startPercent": 100.0,
		"endPercent":   100.0,
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if !eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"}).Enabled {
		t.Error("expected gradual match with 100% range")
	}
}

func TestEvaluate_Gradual_Midpoint(t *testing.T) {
	now := time.Now().UTC()
    // 1h ago to 1h future, 0% to 100%
	rule := ruleWith(models.RuleTypeGradual, map[string]any{
		"startAt":      now.Add(-1 * time.Hour).Format(time.RFC3339),
		"endAt":        now.Add(1 * time.Hour).Format(time.RFC3339),
		"startPercent": 0.0,
		"endPercent":   100.0,
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
    // user-2 bucket is ~15.6. Midpoint is 50.0%. Should match.
	if !eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-2"}).Enabled {
		t.Error("expected gradual match at midpoint")
	}
}

// --- Attribute ---

func TestEvaluate_Attribute_SafeTypes(t *testing.T) {
	tests := []struct {
		name      string
		op        string
		cfgValue  any
		attrValue any
		wantMatch bool
	}{
		{"float match", "gt", "10", 20.0, true},
		{"bool match", "eq", "true", true, true},
		{"string case match", "eq", "PRO", "pro", true},
		{"contains comma match", "contains", "admin", "user, admin", true},
		{"contains substring match", "contains", "foo", "foobar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := ruleWith(models.RuleTypeAttribute, map[string]any{
				"attributeKey":   "k",
				"attributeOp":    tt.op,
				"attributeValue": tt.cfgValue,
			}, true)
			flag := flagWith([]models.Rule{rule}, false, true)
			ctx := models.EvaluationContext{
				Attributes: map[string]any{"k": tt.attrValue},
			}
			if eval.Evaluate(&flag, ctx).Enabled != tt.wantMatch {
				t.Errorf("%s failed", tt.name)
			}
		})
	}
}

func TestEvaluate_StrategyAll(t *testing.T) {
	rule1 := ruleWith(models.RuleTypeUserList, map[string]any{"userIds": []any{"u1"}}, true)
	rule2 := ruleWith(models.RuleTypeGeography, map[string]any{"countries": []any{"US"}}, true)
	flag := flagWith([]models.Rule{rule1, rule2}, false, true)
	flag.RuleMatchStrategy = models.RuleMatchStrategyAll
	
	if !eval.Evaluate(&flag, models.EvaluationContext{UserID: "u1", Country: "US"}).Enabled {
		t.Error("expected ALL strategy match")
	}
}

func TestToString_Extended(t *testing.T) {
	if toString(123.45) != "123.45" { t.Error("float conversion fail") }
    if toString(int32(10)) != "10" { t.Error("int32 conversion fail") }
    if toString(int64(100)) != "100" { t.Error("int64 conversion fail") }
	if toString(true) != "true" { t.Error("bool conversion fail") }
	if toString(nil) != "" { t.Error("nil conversion fail") }
}

func TestToSafeFloat_Extended(t *testing.T) {
    if f, _ := toSafeFloat(int32(10)); f != 10.0 { t.Error("int32 fail") }
    if f, _ := toSafeFloat("12.5"); f != 12.5 { t.Error("string fail") }
}

func TestToSafeTime_Extended(t *testing.T) {
    now := time.Now().UTC()
    if res, _ := toSafeTime(now); !res.Equal(now) { t.Error("time.Time fail") }
    if res, _ := toSafeTime(bson.DateTime(now.UnixMilli())); res.UnixMilli() != now.UnixMilli() { t.Error("bson.DateTime fail") }
}

func TestEvaluate_Gradual_EdgeCases(t *testing.T) {
	now := time.Now().UTC()
    // Before window
	rule1 := ruleWith(models.RuleTypeGradual, map[string]any{
		"startAt":      now.Add(1 * time.Hour).Format(time.RFC3339),
		"endAt":        now.Add(2 * time.Hour).Format(time.RFC3339),
		"startPercent": 0.0,
		"endPercent":   100.0,
	}, true)
	f1 := flagWith([]models.Rule{rule1}, false, true)
	if eval.Evaluate(&f1, models.EvaluationContext{UserID: "u1"}).Enabled {
		t.Error("expected fail before gradual window")
	}
    // After window
	rule2 := ruleWith(models.RuleTypeGradual, map[string]any{
		"startAt":      now.Add(-2 * time.Hour).Format(time.RFC3339),
		"endAt":        now.Add(-1 * time.Hour).Format(time.RFC3339),
		"startPercent": 0.0,
		"endPercent":   0.0,
	}, true)
	f2 := flagWith([]models.Rule{rule2}, false, true)
	if eval.Evaluate(&f2, models.EvaluationContext{UserID: "u1"}).Enabled {
		t.Error("expected fail after gradual window (0%)")
	}
    // Zero duration
	rule3 := ruleWith(models.RuleTypeGradual, map[string]any{
		"startAt":      now.Format(time.RFC3339),
		"endAt":        now.Format(time.RFC3339),
		"startPercent": 0.0,
		"endPercent":   100.0,
	}, true)
	f3 := flagWith([]models.Rule{rule3}, false, true)
	if !eval.Evaluate(&f3, models.EvaluationContext{UserID: "u1"}).Enabled {
		t.Error("expected match for zero duration (returns endPercent)")
	}
}

func TestToStringSlice_Single(t *testing.T) {
    res := toStringSlice(123)
    if len(res) != 1 || res[0] != "123" {
        t.Errorf("expected [123], got %v", res)
    }
    if toStringSlice([]string{"a"})[0] != "a" { t.Error("expected identity for string slice") }
}

func TestEvaluate_Attribute_StringFallback(t *testing.T) {
    // Non-numeric comparison that succeeds as strings
    rule := ruleWith(models.RuleTypeAttribute, map[string]any{
        "attributeKey": "k", "attributeOp": "gt", "attributeValue": "abc",
    }, true)
	f := flagWith([]models.Rule{rule}, false, true)
    // xyz > abc is true
    if !eval.Evaluate(&f, models.EvaluationContext{Attributes: map[string]any{"k": "xyz"}}).Enabled {
        t.Error("expected success for string gt comparison")
    }
}

func TestEvaluate_Gradual_ParsingErrors(t *testing.T) {
    // Missing fields
	rule1 := ruleWith(models.RuleTypeGradual, map[string]any{"startAt": "invalid"}, true)
	f1 := flagWith([]models.Rule{rule1}, false, true)
	if eval.Evaluate(&f1, models.EvaluationContext{UserID: "u1"}).Enabled {
		t.Error("expected fail for invalid date")
	}
}

func TestEvaluate_Schedule_EdgeCases(t *testing.T) {
	now := time.Now().UTC()
    // Invalid range
	rule1 := ruleWith(models.RuleTypeSchedule, map[string]any{
		"enableAt":  now.Add(1 * time.Hour).Format(time.RFC3339),
		"disableAt": now.Add(-1 * time.Hour).Format(time.RFC3339),
	}, true)
	f1 := flagWith([]models.Rule{rule1}, false, true)
	if eval.Evaluate(&f1, models.EvaluationContext{}).Enabled {
		t.Error("expected fail for invalid schedule range")
	}
    // DisableAt only
	rule2 := ruleWith(models.RuleTypeSchedule, map[string]any{
		"disableAt": now.Add(1 * time.Hour).Format(time.RFC3339),
	}, true)
	f2 := flagWith([]models.Rule{rule2}, false, true)
	if !eval.Evaluate(&f2, models.EvaluationContext{}).Enabled {
		t.Error("expected match for future disableAt only")
	}
}

func TestEvaluate_Percentage_MissingUserID(t *testing.T) {
	rule := ruleWith(models.RuleTypePercentage, map[string]any{"percentage": 100.0}, true)
	f := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&f, models.EvaluationContext{UserID: ""}).Enabled {
		t.Error("expected fail for missing userID")
	}
}

func TestEvaluate_MultipleMatches_FirstWins(t *testing.T) {
	rule1 := ruleWith(models.RuleTypeUserList, map[string]any{"userIds": []any{"user-1"}}, true)
	rule2 := ruleWith(models.RuleTypeUserList, map[string]any{"userIds": []any{"user-1"}}, false)
	flag := flagWith([]models.Rule{rule1, rule2}, false, true)
	
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"})
	if !result.Enabled || result.Reason != "matched rule: user_list" {
		t.Errorf("expected first rule to win, got %v", result)
	}
}

func TestEvaluate_StrategyAll_Failure(t *testing.T) {
	rule1 := ruleWith(models.RuleTypeUserList, map[string]any{"userIds": []any{"u1"}}, true)
	rule2 := ruleWith(models.RuleTypeUserList, map[string]any{"userIds": []any{"u2"}}, true)
	flag := flagWith([]models.Rule{rule1, rule2}, false, true)
	flag.RuleMatchStrategy = models.RuleMatchStrategyAll
	
	// u1 matches rule1 but fails rule2
	if eval.Evaluate(&flag, models.EvaluationContext{UserID: "u1"}).Enabled {
		t.Error("expected ALL to fail")
	}
}

func TestEvaluate_UnknownRuleType(t *testing.T) {
	rule := ruleWith("unknown", nil, true)
	flag := flagWith([]models.Rule{rule}, false, true) // defaultValue = false
	if eval.Evaluate(&flag, models.EvaluationContext{}).Enabled {
		t.Error("expected unknown rule to not match")
	}
}

func TestEvaluate_Geography_EmptyConfig(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, map[string]any{}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{Country: "US"}).Enabled {
		t.Error("expected empty geography to fail")
	}
}

func TestEvaluate_Attribute_UnknownOp(t *testing.T) {
	rule := ruleWith(models.RuleTypeAttribute, map[string]any{
		"attributeKey": "k", "attributeOp": "invalid", "attributeValue": "v",
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{Attributes: map[string]any{"k": "v"}}).Enabled {
		t.Error("expected unknown attribute op to fail")
	}
}

func TestEvaluate_Attribute_MissingKey(t *testing.T) {
	rule := ruleWith(models.RuleTypeAttribute, map[string]any{
		"attributeKey": "k", "attributeOp": "eq", "attributeValue": "v",
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{Attributes: map[string]any{"other": "v"}}).Enabled {
		t.Error("expected missing attribute key to fail")
	}
}

func TestEvaluate_Attribute_NumericFail(t *testing.T) {
    rule := ruleWith(models.RuleTypeAttribute, map[string]any{
        "attributeKey": "k", "attributeOp": "gt", "attributeValue": "10",
    }, true)
    flag := flagWith([]models.Rule{rule}, false, true)
    if eval.Evaluate(&flag, models.EvaluationContext{Attributes: map[string]any{"k": nil}}).Enabled {
        t.Error("expected nil attribute to fail numeric check")
    }
}

func TestGetConfig_CaseInsensitive(t *testing.T) {
    m := map[string]any{"UserIds": []any{"u1"}}
    v := getConfig(m, "userIds")
    if v == nil {
        t.Error("expected to find UserIds via case-insensitive lookup")
    }
}

func TestToStringSlice_SingleString(t *testing.T) {
    s := toStringSlice("just-one")
    if len(s) != 1 || s[0] != "just-one" {
        t.Errorf("expected [just-one], got %v", s)
    }
}
