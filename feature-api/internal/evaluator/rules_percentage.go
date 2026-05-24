package evaluator

import (
	"hash"
	"hash/fnv"
	"sync"
	"time"

	"github.com/featureflags/feature-api/internal/models"
)

var fnvPool = sync.Pool{
	New: func() any {
		return fnv.New64a()
	},
}

func getBucket(flagKey, userID string) float64 {
	h := fnvPool.Get().(hash.Hash64)
	defer fnvPool.Put(h)
	h.Reset()
	h.Write([]byte(flagKey + ":" + userID))
	// Use 10000 for 0.01% precision
	return float64(h.Sum64()%10000) / 100.0
}

func evalPercentageRule(rule models.Rule, flagKey string, ctx models.EvaluationContext) (bool, bool) {
	if rule.Config.Percentage == nil || ctx.UserID == "" {
		return false, false
	}

	bucket := getBucket(flagKey, ctx.UserID)
	return bucket < *rule.Config.Percentage, rule.Value
}

func evalGradualRule(rule models.Rule, flagKey string, ctx models.EvaluationContext, now time.Time) (bool, bool) {
	if ctx.UserID == "" {
		return false, false
	}

	c := rule.Config
	if c.StartPercent == nil || c.EndPercent == nil || c.StartAt == nil || c.EndAt == nil {
		return false, false
	}

	var effectivePercent float64
	if now.Before(*c.StartAt) {
		effectivePercent = *c.StartPercent
	} else if now.After(*c.EndAt) {
		effectivePercent = *c.EndPercent
	} else {
		duration := c.EndAt.Sub(*c.StartAt)
		if duration <= 0 {
			effectivePercent = *c.EndPercent
		} else {
			progress := float64(now.Sub(*c.StartAt)) / float64(duration)
			effectivePercent = *c.StartPercent + progress*(*c.EndPercent-*c.StartPercent)
		}
	}

	bucket := getBucket(flagKey, ctx.UserID)
	return bucket < effectivePercent, rule.Value
}
