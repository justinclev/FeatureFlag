package evaluator

import (
	"time"

	"github.com/featureflags/feature-api/internal/models"
)

func evalScheduleRule(rule models.Rule, now time.Time) (bool, bool) {
	if len(rule.Config) == 0 {
		return false, false
	}

	enableAt, ok1 := toSafeTime(rule.Config["enableAt"])
	disableAt, ok2 := toSafeTime(rule.Config["disableAt"])

	if !ok1 && !ok2 {
		return false, false
	}

	// Sanity
	if !enableAt.IsZero() && !disableAt.IsZero() {
		if enableAt.After(disableAt) {
			return false, false
		}
	}

	if !enableAt.IsZero() && now.Before(enableAt) {
		return false, false
	}

	if !disableAt.IsZero() && now.After(disableAt) {
		return false, false
	}

	return true, rule.Value
}
