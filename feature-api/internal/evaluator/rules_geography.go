package evaluator

import (
	"strings"

	"github.com/featureflags/feature-api/internal/models"
)

func evalGeographyRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	if len(rule.Config.Countries) == 0 && len(rule.Config.States) == 0 && len(rule.Config.ZipCodes) == 0 && len(rule.Config.Cities) == 0 {
		return false, false
	}

	if len(rule.Config.Countries) > 0 {
		matched := false
		for _, country := range rule.Config.Countries {
			if strings.EqualFold(country, ctx.Country) {
				matched = true
				break
			}
		}
		if !matched {
			return false, false
		}
	}

	if len(rule.Config.States) > 0 {
		matched := false
		for _, state := range rule.Config.States {
			if strings.EqualFold(state, ctx.State) {
				matched = true
				break
			}
		}
		if !matched {
			return false, false
		}
	}
	if len(rule.Config.Cities) > 0 {
		matched := false
		for _, city := range rule.Config.Cities {
			if strings.EqualFold(city, ctx.City) {
				matched = true
				break
			}
		}
		if !matched {
			return false, false
		}
	}

	if len(rule.Config.ZipCodes) > 0 {
		matched := false
		for _, zip := range rule.Config.ZipCodes {
			if zip == ctx.ZipCode {
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
