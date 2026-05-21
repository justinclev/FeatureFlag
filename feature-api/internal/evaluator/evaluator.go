package evaluator

import (
	"github.com/featureflags/feature-api/internal/models"
)

type Evaluator struct{}

func New() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(flag models.Flag, ctx models.EvaluationContext) models.EvaluationResult {
	if !flag.Enabled {
		return models.EvaluationResult{Enabled: false, Reason: "flag disabled"}
	}

	for _, rule := range flag.Rules {
		if matched, value := evalRule(rule, flag.Key, ctx); matched {
			return models.EvaluationResult{Enabled: value, Reason: "matched rule: " + string(rule.Type)}
		}
	}

	return models.EvaluationResult{Enabled: flag.DefaultValue, Reason: "default value"}
}

func evalRule(rule models.Rule, flagKey string, ctx models.EvaluationContext) (bool, bool) {
	switch rule.Type {
	case models.RuleTypePercentage:
		return evalPercentageRule(rule, flagKey, ctx)
	case models.RuleTypeGeography:
		return evalGeographyRule(rule, ctx)
	case models.RuleTypeSchedule:
		return evalScheduleRule(rule)
	case models.RuleTypeGradual:
		return evalGradualRule(rule, flagKey, ctx)
	case models.RuleTypeUserList:
		return evalUserListRule(rule, ctx)
	case models.RuleTypeAttribute:
		return evalAttributeRule(rule, ctx)
	default:
		return false, false
	}
}
