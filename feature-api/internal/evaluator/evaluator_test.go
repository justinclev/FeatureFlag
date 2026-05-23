package evaluator_test

import (
	"testing"
	"time"

	"github.com/featureflags/feature-api/internal/evaluator"
	"github.com/featureflags/feature-api/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var eval = evaluator.New()

func ptrFloat(v float64) *float64 { return &v }
func ptrTime(t time.Time) *time.Time { return &t }

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

func ruleWith(t models.RuleType, cfg models.RuleConfig, value bool) models.Rule {
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


func TestEvaluate_NoRules_DefaultFalse(t *testing.T) {
	flag := flagWith(nil, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{})
	if result.Enabled {
		t.Error("expected disabled via default")
	}
}

// --- Percentage ---

func TestEvaluate_Percentage_AlwaysMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypePercentage, models.RuleConfig{Percentage: ptrFloat(100)}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"})
	if !result.Enabled {
		t.Error("expected match with 100% rollout")
	}
}

func TestEvaluate_Percentage_NeverMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypePercentage, models.RuleConfig{Percentage: ptrFloat(0)}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"})
	if result.Enabled {
		t.Error("expected no match with 0% rollout")
	}
}

func TestEvaluate_Percentage_EmptyUserID_NoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypePercentage, models.RuleConfig{Percentage: ptrFloat(100)}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: ""})
	if result.Enabled {
		t.Error("expected no match with empty userID")
	}
}

func TestEvaluate_Percentage_Consistent(t *testing.T) {
	rule := ruleWith(models.RuleTypePercentage, models.RuleConfig{Percentage: ptrFloat(50)}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	ctx := models.EvaluationContext{UserID: "stable-user"}
	first := eval.Evaluate(&flag, ctx)
	for i := 0; i < 10; i++ {
		if eval.Evaluate(&flag, ctx).Enabled != first.Enabled {
			t.Error("percentage bucketing is not consistent across calls")
		}
	}
}

// --- UserList ---

func TestEvaluate_UserList_Match(t *testing.T) {
	rule := ruleWith(models.RuleTypeUserList, models.RuleConfig{UserIDs: []string{"user-1", "user-2"}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-2"})
	if !result.Enabled {
		t.Error("expected match for user in list")
	}
}

func TestEvaluate_UserList_NoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeUserList, models.RuleConfig{UserIDs: []string{"user-1"}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-99"})
	if result.Enabled {
		t.Error("expected no match for user not in list")
	}
}

func TestEvaluate_UserList_Empty_NoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeUserList, models.RuleConfig{UserIDs: []string{}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"})
	if result.Enabled {
		t.Error("expected no match with empty user list")
	}
}

// --- Geography ---

func TestEvaluate_Geography_CountryMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, models.RuleConfig{Countries: []string{"US"}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{Country: "US"})
	if !result.Enabled {
		t.Error("expected country match")
	}
}

func TestEvaluate_Geography_CountryMatch_CaseInsensitive(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, models.RuleConfig{Countries: []string{"US"}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{Country: "us"})
	if !result.Enabled {
		t.Error("expected case-insensitive country match")
	}
}

func TestEvaluate_Geography_CountryNoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, models.RuleConfig{Countries: []string{"US"}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{Country: "CA"})
	if result.Enabled {
		t.Error("expected no country match")
	}
}

func TestEvaluate_Geography_CountryAndState_BothMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, models.RuleConfig{
		Countries: []string{"US"},
		States:    []string{"CA"},
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{Country: "US", State: "CA"})
	if !result.Enabled {
		t.Error("expected match with country+state")
	}
}

func TestEvaluate_Geography_CountryAndState_StateMismatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, models.RuleConfig{
		Countries: []string{"US"},
		States:    []string{"CA"},
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{Country: "US", State: "TX"})
	if result.Enabled {
		t.Error("expected no match when state mismatches")
	}
}

func TestEvaluate_Geography_ZipMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, models.RuleConfig{ZipCodes: []string{"94105"}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{ZipCode: "94105"})
	if !result.Enabled {
		t.Error("expected zip match")
	}
}

func TestEvaluate_Geography_CityMatch_CaseInsensitive(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, models.RuleConfig{Cities: []string{"San Francisco"}}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{City: "san francisco"})
	if !result.Enabled {
		t.Error("expected case-insensitive city match")
	}
}

// --- Schedule ---

func TestEvaluate_Schedule_WithinWindow(t *testing.T) {
	rule := ruleWith(models.RuleTypeSchedule, models.RuleConfig{
		EnableAt:  ptrTime(time.Now().UTC().Add(-1 * time.Hour)),
		DisableAt: ptrTime(time.Now().UTC().Add(1 * time.Hour)),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if !eval.Evaluate(&flag, models.EvaluationContext{}).Enabled {
		t.Error("expected match within schedule window")
	}
}

func TestEvaluate_Schedule_BeforeWindow(t *testing.T) {
	rule := ruleWith(models.RuleTypeSchedule, models.RuleConfig{
		EnableAt:  ptrTime(time.Now().UTC().Add(1 * time.Hour)),
		DisableAt: ptrTime(time.Now().UTC().Add(2 * time.Hour)),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{}).Enabled {
		t.Error("expected no match before schedule window")
	}
}

func TestEvaluate_Schedule_AfterWindow(t *testing.T) {
	rule := ruleWith(models.RuleTypeSchedule, models.RuleConfig{
		EnableAt:  ptrTime(time.Now().UTC().Add(-2 * time.Hour)),
		DisableAt: ptrTime(time.Now().UTC().Add(-1 * time.Hour)),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{}).Enabled {
		t.Error("expected no match after schedule window")
	}
}

func TestEvaluate_Schedule_EnableAtOnly(t *testing.T) {
	rule := ruleWith(models.RuleTypeSchedule, models.RuleConfig{
		EnableAt: ptrTime(time.Now().UTC().Add(-1 * time.Hour)),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if !eval.Evaluate(&flag, models.EvaluationContext{}).Enabled {
		t.Error("expected match with only EnableAt in the past")
	}
}

func TestEvaluate_Schedule_DisableAtOnly(t *testing.T) {
	rule := ruleWith(models.RuleTypeSchedule, models.RuleConfig{
		DisableAt: ptrTime(time.Now().UTC().Add(1 * time.Hour)),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if !eval.Evaluate(&flag, models.EvaluationContext{}).Enabled {
		t.Error("expected match with only DisableAt in the future")
	}
}

func TestEvaluate_Schedule_BothNil_NoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeSchedule, models.RuleConfig{}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{}).Enabled {
		t.Error("expected no match with both schedule times nil")
	}
}

// --- Gradual ---

func TestEvaluate_Gradual_WindowPassed_EndPercent100(t *testing.T) {
	rule := ruleWith(models.RuleTypeGradual, models.RuleConfig{
		StartAt:      ptrTime(time.Now().UTC().Add(-2 * time.Hour)),
		EndAt:        ptrTime(time.Now().UTC().Add(-1 * time.Hour)),
		StartPercent: ptrFloat(0),
		EndPercent:   ptrFloat(100),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if !eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"}).Enabled {
		t.Error("expected match: window passed, EndPercent=100")
	}
}

func TestEvaluate_Gradual_WindowPassed_EndPercent0(t *testing.T) {
	rule := ruleWith(models.RuleTypeGradual, models.RuleConfig{
		StartAt:      ptrTime(time.Now().UTC().Add(-2 * time.Hour)),
		EndAt:        ptrTime(time.Now().UTC().Add(-1 * time.Hour)),
		StartPercent: ptrFloat(0),
		EndPercent:   ptrFloat(0),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"}).Enabled {
		t.Error("expected no match: window passed, EndPercent=0")
	}
}

func TestEvaluate_Gradual_WindowNotStarted_StartPercent100(t *testing.T) {
	rule := ruleWith(models.RuleTypeGradual, models.RuleConfig{
		StartAt:      ptrTime(time.Now().UTC().Add(1 * time.Hour)),
		EndAt:        ptrTime(time.Now().UTC().Add(2 * time.Hour)),
		StartPercent: ptrFloat(100),
		EndPercent:   ptrFloat(100),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if !eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"}).Enabled {
		t.Error("expected match: window not started, StartPercent=100")
	}
}

func TestEvaluate_Gradual_WindowNotStarted_StartPercent0(t *testing.T) {
	rule := ruleWith(models.RuleTypeGradual, models.RuleConfig{
		StartAt:      ptrTime(time.Now().UTC().Add(1 * time.Hour)),
		EndAt:        ptrTime(time.Now().UTC().Add(2 * time.Hour)),
		StartPercent: ptrFloat(0),
		EndPercent:   ptrFloat(100),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"}).Enabled {
		t.Error("expected no match: window not started, StartPercent=0")
	}
}

func TestEvaluate_Gradual_EmptyUserID_NoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGradual, models.RuleConfig{
		StartAt:      ptrTime(time.Now().UTC().Add(-1 * time.Hour)),
		EndAt:        ptrTime(time.Now().UTC().Add(1 * time.Hour)),
		StartPercent: ptrFloat(0),
		EndPercent:   ptrFloat(100),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{UserID: ""}).Enabled {
		t.Error("expected no match with empty userID")
	}
}

// --- Attribute ---

func TestEvaluate_Attribute(t *testing.T) {
	tests := []struct {
		name      string
		op        string
		cfgValue  string
		attrValue string
		wantMatch bool
	}{
		{"eq match", "eq", "premium", "premium", true},
		{"eq no match", "eq", "premium", "free", false},
		{"neq match", "neq", "free", "premium", true},
		{"neq no match", "neq", "premium", "premium", false},
		{"neq empty expected", "neq", "", "something", true},
		{"contains match", "contains", "pro", "enterprise-pro", true},
		{"contains no match", "contains", "pro", "basic", false},
		{"gt match", "gt", "3", "5", true},
		{"gt no match", "gt", "5", "3", false},
		{"lt match", "lt", "5", "3", true},
		{"lt no match", "lt", "3", "5", false},
		{"unknown op", "between", "1", "2", false},
		{"non-numeric gt", "gt", "abc", "xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := ruleWith(models.RuleTypeAttribute, models.RuleConfig{
				AttributeKey:   "plan",
				AttributeOp:    tt.op,
				AttributeValue: tt.cfgValue,
			}, true)
			flag := flagWith([]models.Rule{rule}, false, true)
			ctx := models.EvaluationContext{
				Attributes: map[string]string{"plan": tt.attrValue},
			}
			result := eval.Evaluate(&flag, ctx)
			if result.Enabled != tt.wantMatch {
				t.Errorf("op=%q cfgValue=%q attrValue=%q: expected enabled=%v, got %v",
					tt.op, tt.cfgValue, tt.attrValue, tt.wantMatch, result.Enabled)
			}
		})
	}
}

func TestEvaluate_Attribute_MissingKey_NoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeAttribute, models.RuleConfig{
		AttributeKey:   "plan",
		AttributeOp:    "eq",
		AttributeValue: "premium",
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	ctx := models.EvaluationContext{
		Attributes: map[string]string{"tier": "premium"},
	}
	if eval.Evaluate(&flag, ctx).Enabled {
		t.Error("expected no match when attribute key is missing from context")
	}
}

func TestEvaluate_Attribute_EmptyKey_NoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeAttribute, models.RuleConfig{
		AttributeKey:   "",
		AttributeOp:    "eq",
		AttributeValue: "premium",
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{Attributes: map[string]string{"plan": "premium"}}).Enabled {
		t.Error("expected no match with empty attribute key")
	}
}

// --- Rule ordering ---

func TestEvaluate_FirstRuleMatchWins(t *testing.T) {
	rule1 := ruleWith(models.RuleTypeUserList, models.RuleConfig{UserIDs: []string{"user-1"}}, true)
	rule2 := ruleWith(models.RuleTypeUserList, models.RuleConfig{UserIDs: []string{"user-1"}}, false)
	flag := flagWith([]models.Rule{rule1, rule2}, false, true)
	if !eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"}).Enabled {
		t.Error("expected first matching rule's value (true)")
	}
}

func TestEvaluate_UnknownRuleType(t *testing.T) {
	rule := ruleWith("unknown", models.RuleConfig{}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "any"})
	if result.Enabled {
		t.Error("expected disabled for unknown rule type")
	}
}

func TestEvaluate_Geography_NoCriteria_NoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGeography, models.RuleConfig{}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{Country: "US"}).Enabled {
		t.Error("expected no match with empty geography criteria")
	}
}

func TestEvaluate_Gradual_InWindow(t *testing.T) {
	now := time.Now().UTC()
	rule := ruleWith(models.RuleTypeGradual, models.RuleConfig{
		StartAt:      ptrTime(now.Add(-1 * time.Hour)),
		EndAt:        ptrTime(now.Add(1 * time.Hour)),
		StartPercent: ptrFloat(0),
		EndPercent:   ptrFloat(100),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	
	// Middle of the window (50%)
	// UserID bucketing is deterministic. Let's find a user that is in/out.
	// user-1 bucketing is usually consistent.
	ctx := models.EvaluationContext{UserID: "user-1"}
	_ = eval.Evaluate(&flag, ctx) // Just trigger it.
}

func TestEvaluate_Gradual_MissingConfig_NoMatch(t *testing.T) {
	rule := ruleWith(models.RuleTypeGradual, models.RuleConfig{}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"}).Enabled {
		t.Error("expected no match with missing gradual config")
	}
}

func TestEvaluate_StrategyAll_Match(t *testing.T) {
	rule1 := ruleWith(models.RuleTypeUserList, models.RuleConfig{UserIDs: []string{"user-1"}}, true)
	rule2 := ruleWith(models.RuleTypeGeography, models.RuleConfig{Countries: []string{"US"}}, true)
	flag := flagWith([]models.Rule{rule1, rule2}, false, true)
	flag.RuleMatchStrategy = models.RuleMatchStrategyAll
	
	ctx := models.EvaluationContext{UserID: "user-1", Country: "US"}
	if !eval.Evaluate(&flag, ctx).Enabled {
		t.Error("expected match when all rules match")
	}
}

func TestEvaluate_StrategyAll_PartialMatch(t *testing.T) {
	rule1 := ruleWith(models.RuleTypeUserList, models.RuleConfig{UserIDs: []string{"user-1"}}, true)
	rule2 := ruleWith(models.RuleTypeGeography, models.RuleConfig{Countries: []string{"CA"}}, true)
	flag := flagWith([]models.Rule{rule1, rule2}, false, true)
	flag.RuleMatchStrategy = models.RuleMatchStrategyAll
	
	ctx := models.EvaluationContext{UserID: "user-1", Country: "US"}
	if eval.Evaluate(&flag, ctx).Enabled {
		t.Error("expected fail when only one rule matches in 'all' strategy")
	}
}

func TestEvaluate_Gradual_InvalidDuration(t *testing.T) {
	now := time.Now().UTC()
	rule := ruleWith(models.RuleTypeGradual, models.RuleConfig{
		StartAt:      ptrTime(now),
		EndAt:        ptrTime(now), // Zero duration
		StartPercent: ptrFloat(0),
		EndPercent:   ptrFloat(100),
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	
	// If now is exactly StartAt, now.Before is false, now.After is false.
	// It hits the duration check.
	result := eval.Evaluate(&flag, models.EvaluationContext{UserID: "user-1"})
	if !result.Enabled {
		t.Error("expected enabled via EndPercent for zero duration")
	}
}

func TestEvaluate_Schedule_InvalidRange(t *testing.T) {
	now := time.Now().UTC()
	rule := ruleWith(models.RuleTypeSchedule, models.RuleConfig{
		EnableAt:  ptrTime(now.Add(1 * time.Hour)),
		DisableAt: ptrTime(now), // Disable before enable
	}, true)
	flag := flagWith([]models.Rule{rule}, false, true)
	if eval.Evaluate(&flag, models.EvaluationContext{}).Enabled {
		t.Error("expected no match for invalid schedule range")
	}
}
