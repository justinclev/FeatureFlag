package evaluator

import (
	"hash"
	"hash/fnv"
	"io"
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
	_, _ = io.WriteString(h, flagKey)
	_, _ = io.WriteString(h, ":")
	_, _ = io.WriteString(h, userID)
	return float64(h.Sum64()%10000) / 100.0
}

func evalPercentageRule(rule models.Rule, flagKey string, ctx models.EvaluationContext) (bool, bool) {
	p, ok := rule.Config["percentage"].(float64)
	if !ok || ctx.UserID == "" {
		return false, false
	}

	bucket := getBucket(flagKey, ctx.UserID)
	return bucket < p, rule.Value
}

func evalGradualRule(rule models.Rule, flagKey string, ctx models.EvaluationContext, now time.Time) (bool, bool) {
	if ctx.UserID == "" {
		return false, false
	}

	sp, ok1 := rule.Config["startPercent"].(float64)
	ep, ok2 := rule.Config["endPercent"].(float64)
	saRaw, ok3 := rule.Config["startAt"].(string)
	eaRaw, ok4 := rule.Config["endAt"].(string)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return false, false
	}

	startAt, err1 := time.Parse(time.RFC3339, saRaw)
	endAt, err2 := time.Parse(time.RFC3339, eaRaw)
	if err1 != nil || err2 != nil {
		return false, false
	}

	var effectivePercent float64
	if now.Before(startAt) {
		effectivePercent = sp
	} else if now.After(endAt) {
		effectivePercent = ep
	} else {
		duration := endAt.Sub(startAt)
		if duration <= 0 {
			effectivePercent = ep
		} else {
			progress := float64(now.Sub(startAt)) / float64(duration)
			effectivePercent = sp + progress*(ep-sp)
		}
	}

	bucket := getBucket(flagKey, ctx.UserID)
	return bucket < effectivePercent, rule.Value
}
