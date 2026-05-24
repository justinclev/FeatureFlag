package evaluator

import (
	"time"

	"github.com/featureflags/feature-api/internal/models"
)

func evalScheduleRule(rule models.Rule, now time.Time) (bool, bool) {
	if rule.Config.EnableAt == nil && rule.Config.DisableAt == nil {
		return false, false
	}

	// Sanity: if both exist, EnableAt must be before DisableAt
	if rule.Config.EnableAt != nil && rule.Config.DisableAt != nil {
		if rule.Config.EnableAt.After(*rule.Config.DisableAt) {
			return false, false
		}
	}

	if rule.Config.EnableAt != nil && now.Before(*rule.Config.EnableAt) {
		return false, false
	}

	if rule.Config.DisableAt != nil && now.After(*rule.Config.DisableAt) {
		return false, false
	}

	return true, rule.Value
}
