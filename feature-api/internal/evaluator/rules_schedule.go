package evaluator

import (
	"time"

	"github.com/featureflags/feature-api/internal/models"
)

func evalScheduleRule(rule models.Rule, now time.Time) (bool, bool) {
	if len(rule.Config) == 0 {
		return false, false
	}

	eaRaw, ok1 := rule.Config["enableAt"].(string)
	daRaw, ok2 := rule.Config["disableAt"].(string)

	if !ok1 && !ok2 {
		return false, false
	}

	var enableAt, disableAt time.Time
	if ok1 && eaRaw != "" {
		enableAt, _ = time.Parse(time.RFC3339, eaRaw)
	}
	if ok2 && daRaw != "" {
		disableAt, _ = time.Parse(time.RFC3339, daRaw)
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
