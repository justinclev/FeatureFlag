package evaluator

import (
	"strconv"
	"strings"

	"github.com/featureflags/feature-api/internal/models"
)

func evalAttributeRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	akRaw, ok1 := rule.Config["attributeKey"].(string)
	aoRaw, ok2 := rule.Config["attributeOp"].(string)
	avRaw := toString(rule.Config["attributeValue"])

	if !ok1 || !ok2 || akRaw == "" || aoRaw == "" {
		return false, false
	}

	rawActual, ok := ctx.Attributes[akRaw]
	if !ok {
		return false, false
	}

	// Principal refinement: Unified case-insensitive and trimmed comparison.
	actual := strings.ToLower(strings.TrimSpace(toString(rawActual)))
	expected := strings.ToLower(strings.TrimSpace(avRaw))
	var matched bool

	switch aoRaw {
	case "eq":
		matched = (actual == expected)
	case "neq":
		matched = (actual != expected)
	case "contains":
		// Check for comma-separated list match first (resilient to spacing)
		if strings.Contains(actual, ",") {
			parts := strings.Split(actual, ",")
			for _, p := range parts {
				if strings.TrimSpace(p) == expected {
					matched = true
					break
				}
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
		if aoRaw == "gt" {
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
