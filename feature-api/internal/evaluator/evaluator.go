package evaluator

import (
	"time"

	"github.com/featureflags/feature-api/internal/models"
)

// Evaluator provides logic to determine if a flag is enabled for a given context.
type Evaluator struct{}

// New creates a new Evaluator.
func New() *Evaluator {
	return &Evaluator{}
}

// Evaluate determines the enabled state of a flag based on its rules and the provided context.
func (e *Evaluator) Evaluate(flag *models.Flag, ctx models.EvaluationContext) models.EvaluationResult {
	// Capture time once for the entire evaluation.
	now := time.Now().UTC()

	if !flag.Enabled {
		return models.EvaluationResult{Enabled: false, Reason: "flag disabled"}
	}

	if len(flag.Rules) == 0 {
		return models.EvaluationResult{Enabled: flag.DefaultValue, Reason: "default value (no rules)"}
	}

	var result models.EvaluationResult
	switch flag.RuleMatchStrategy {
	case models.RuleMatchStrategyAll:
		result = e.evaluateAll(flag, ctx, now)
	default: // RuleMatchStrategyAny is default
		result = e.evaluateAny(flag, ctx, now)
	}

	return result
}

func (e *Evaluator) evaluateAny(flag *models.Flag, ctx models.EvaluationContext, now time.Time) models.EvaluationResult {
	hasTrueMatch := false
	for _, rule := range flag.Rules {
		matched, value := evalRule(rule, flag.Key, ctx, now)
		if matched {
			if !value {
				// Deny Wins: immediately return false if any matching rule is false
				return models.EvaluationResult{Enabled: false, Reason: "matched rule (deny): " + string(rule.Type)}
			}
			hasTrueMatch = true
		}
	}
	if hasTrueMatch {
		return models.EvaluationResult{Enabled: true, Reason: "matched rule (permit)"}
	}
	return models.EvaluationResult{Enabled: flag.DefaultValue, Reason: "default value"}
}

func (e *Evaluator) evaluateAll(flag *models.Flag, ctx models.EvaluationContext, now time.Time) models.EvaluationResult {
	if len(flag.Rules) == 0 {
		return models.EvaluationResult{Enabled: flag.DefaultValue, Reason: "default value (no rules)"}
	}

	for _, rule := range flag.Rules {
		matched, _ := evalRule(rule, flag.Key, ctx, now)
		if !matched {
			return models.EvaluationResult{Enabled: flag.DefaultValue, Reason: "failed rule: " + string(rule.Type)}
		}
	}
	// All matched. Since validation ensures all rules have the same value for 'all' strategy,
	// we can return the value of the first rule.
	return models.EvaluationResult{Enabled: flag.Rules[0].Value, Reason: "matched all rules"}
}

func evalRule(rule models.Rule, flagKey string, ctx models.EvaluationContext, now time.Time) (bool, bool) {
	switch rule.Type {
	case models.RuleTypePercentage:
		return evalPercentageRule(rule, flagKey, ctx)
	case models.RuleTypeGeography:
		return evalGeographyRule(rule, ctx)
	case models.RuleTypeSchedule:
		return evalScheduleRule(rule, now)
	case models.RuleTypeGradual:
		return evalGradualRule(rule, flagKey, ctx, now)
	case models.RuleTypeUserList:
		return evalUserListRule(rule, ctx)
	case models.RuleTypeAttribute:
		return evalAttributeRule(rule, ctx)
	default:
		return false, false
	}
}
