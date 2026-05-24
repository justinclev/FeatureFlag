package evaluator

import (
	"strings"
	"sync"

	"github.com/featureflags/feature-api/internal/models"
)

var compiledGeography sync.Map // map[string]*geographySets

type geographySets struct {
	countries map[string]struct{}
	states    map[string]struct{}
	cities    map[string]struct{}
	zipCodes  map[string]struct{}
}

func evalGeographyRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	if len(rule.Config) == 0 {
		return false, false
	}

	ruleID := rule.ID.Hex()
	var sets *geographySets

	// Fast path: load compiled sets
	if cached, ok := compiledGeography.Load(ruleID); ok {
		sets = cached.(*geographySets)
	} else {
		// Slow path: compile and store
		sets = &geographySets{
			countries: toMap(toStringSlice(getConfig(rule.Config, "countries"))),
			states:    toMap(toStringSlice(getConfig(rule.Config, "states"))),
			cities:    toMap(toStringSlice(getConfig(rule.Config, "cities"))),
			zipCodes:  toMap(toStringSlice(getConfig(rule.Config, "zipCodes"))),
		}
		compiledGeography.Store(ruleID, sets)
	}

	if len(sets.countries) == 0 && len(sets.states) == 0 && len(sets.zipCodes) == 0 && len(sets.cities) == 0 {
		return false, false
	}

	if len(sets.countries) > 0 {
		if _, ok := sets.countries[strings.ToLower(strings.TrimSpace(ctx.Country))]; !ok {
			return false, false
		}
	}

	if len(sets.states) > 0 {
		if _, ok := sets.states[strings.ToLower(strings.TrimSpace(ctx.State))]; !ok {
			return false, false
		}
	}

	if len(sets.cities) > 0 {
		if _, ok := sets.cities[strings.ToLower(strings.TrimSpace(ctx.City))]; !ok {
			return false, false
		}
	}

	if len(sets.zipCodes) > 0 {
		if _, ok := sets.zipCodes[strings.ToLower(strings.TrimSpace(ctx.ZipCode))]; !ok {
			return false, false
		}
	}

	return true, rule.Value
}

func toMap(slice []string) map[string]struct{} {
	if len(slice) == 0 {
		return nil
	}
	m := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		m[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
	}
	return m
}
