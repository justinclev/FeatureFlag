package evaluator

import (
	"strconv"
	"strings"

	"github.com/featureflags/feature-api/internal/models"
)

func evalAttributeRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	akRaw, ok1 := getConfig(rule.Config, "attributeKey").(string)
	aoRaw, ok2 := getConfig(rule.Config, "attributeOp").(string)
	avRaw := toString(getConfig(rule.Config, "attributeValue"))

	if !ok1 || !ok2 || akRaw == "" || aoRaw == "" {
		return false, false
	}

	// Principal optimization: Case-insensitive attribute key lookup
	var rawActual any
	found := false
	for k, v := range ctx.Attributes {
		if stringsEqualFold(k, akRaw) {
			rawActual = v
			found = true
			break
		}
	}
	if !found {
		return false, false
	}

	actual := strings.ToLower(strings.TrimSpace(toString(rawActual)))
	expected := strings.ToLower(strings.TrimSpace(avRaw))
	var matched bool

	switch aoRaw {
	case "eq":
		matched = (actual == expected)
	case "neq":
		matched = (actual != expected)
	case "contains":
		if strings.Contains(actual, ",") {
			parts := strings.Split(actual, ",")
			for _, p := range parts {
				if strings.ToLower(strings.TrimSpace(p)) == expected {
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
		if err1 == nil && err2 == nil {
			if aoRaw == "gt" {
				matched = actualF > expectedF
			} else {
				matched = actualF < expectedF
			}
		} else {
			if aoRaw == "gt" {
				matched = actual > expected
			} else {
				matched = actual < expected
			}
		}
	default:
		return false, false
	}

	if !matched {
		return false, false
	}

	return true, rule.Value
}
