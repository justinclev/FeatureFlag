package evaluator

import (
	"strconv"
	"strings"

	"github.com/featureflags/feature-api/internal/models"
)

func evalAttributeRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	if rule.Config.AttributeKey == "" || rule.Config.AttributeOp == "" {
		return false, false
	}

	actual, ok := ctx.Attributes[rule.Config.AttributeKey]
	if !ok {
		return false, false
	}

	expected := rule.Config.AttributeValue
	var matched bool

	switch rule.Config.AttributeOp {
	case "eq":
		matched = actual == expected
	case "neq":
		matched = actual != expected
	case "contains":
		matched = strings.Contains(actual, expected)
	case "gt", "lt":
		actualF, err1 := strconv.ParseFloat(actual, 64)
		expectedF, err2 := strconv.ParseFloat(expected, 64)
		if err1 != nil || err2 != nil {
			return false, false
		}
		if rule.Config.AttributeOp == "gt" {
			matched = actualF > expectedF
		} else {
			matched = actualF < expectedF
		}
	default:
		return false, false
	}
	if !matched {
		return false, false
	}

	return true, rule.Value
}
