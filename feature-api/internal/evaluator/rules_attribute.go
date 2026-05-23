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
		matched = strings.EqualFold(actual, expected)
	case "neq":
		matched = !strings.EqualFold(actual, expected)
	case "contains":
		// Check if expected is in actual (e.g. actual="admin,user", expected="admin")
		// Or if actual is a comma-separated list
		parts := strings.Split(actual, ",")
		for _, p := range parts {
			if strings.TrimSpace(p) == expected {
				matched = true
				break
			}
		}
		if !matched {
			matched = strings.Contains(actual, expected)
		}
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
