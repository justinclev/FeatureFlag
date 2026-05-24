package evaluator

import (
	"strings"

	"github.com/featureflags/feature-api/internal/models"
)

func evalGeographyRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	if len(rule.Config) == 0 {
		return false, false
	}

	countries := toStringSlice(rule.Config["countries"])
	states := toStringSlice(rule.Config["states"])
	cities := toStringSlice(rule.Config["cities"])
	zipCodes := toStringSlice(rule.Config["zipCodes"])

	if len(countries) == 0 && len(states) == 0 && len(zipCodes) == 0 && len(cities) == 0 {
		return false, false
	}

	if len(countries) > 0 {
		matched := false
		for _, country := range countries {
			if strings.EqualFold(country, ctx.Country) {
				matched = true
				break
			}
		}
		if !matched {
			return false, false
		}
	}

	if len(states) > 0 {
		matched := false
		for _, state := range states {
			if strings.EqualFold(state, ctx.State) {
				matched = true
				break
			}
		}
		if !matched {
			return false, false
		}
	}

	if len(cities) > 0 {
		matched := false
		for _, city := range cities {
			if strings.EqualFold(city, ctx.City) {
				matched = true
				break
			}
		}
		if !matched {
			return false, false
		}
	}

	if len(zipCodes) > 0 {
		matched := false
		for _, zip := range zipCodes {
			if strings.EqualFold(zip, ctx.ZipCode) {
				matched = true
				break
			}
		}
		if !matched {
			return false, false
		}
	}

	return true, rule.Value
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	raw, ok := v.([]any)
	if !ok {
		// Try string slice if it was already converted
		if s, ok := v.([]string); ok {
			return s
		}
		return nil
	}
	res := make([]string, len(raw))
	for i, val := range raw {
		res[i] = toString(val)
	}
	return res
}
