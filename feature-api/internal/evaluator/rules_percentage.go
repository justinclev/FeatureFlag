package evaluator

import (
	"hash/fnv"
	"time"

	"github.com/featureflags/feature-api/internal/models"
)

func evalPercentageRule(rule models.Rule, flagKey string, ctx models.EvaluationContext) (bool, bool) {
	// Implement percentage-based evaluation logic here
	if rule.Config.Percentage == nil || ctx.UserID == "" {
		return false, false
	}

	h := fnv.New64a()
	h.Write([]byte(flagKey + ":" + ctx.UserID))
	bucket := float64(h.Sum64() % 100)

	return bucket < *rule.Config.Percentage, rule.Value
}

func evalGradualRule(rule models.Rule, flagKey string, ctx models.EvaluationContext) (bool, bool) {
	if ctx.UserID == "" {
		return false, false
	}

	c := rule.Config
	if c.StartPercent == nil || c.EndPercent == nil || c.StartAt == nil || c.EndAt == nil {
		return false, false
	}

	now := time.Now().UTC()

	var effectivePercent float64
	if now.Before(*c.StartAt) {
		effectivePercent = *c.StartPercent
	} else if now.After(*c.EndAt) {
		effectivePercent = *c.EndPercent
	} else {
		progress := float64(now.Sub(*c.StartAt)) / float64(c.EndAt.Sub(*c.StartAt))
		effectivePercent = *c.StartPercent + progress*(*c.EndPercent-*c.StartPercent)
	}

	h := fnv.New64a()
	h.Write([]byte(flagKey + ":" + ctx.UserID))
	bucket := float64(h.Sum64() % 100)

	return bucket < effectivePercent, rule.Value
}
