package evaluator

import (
	"github.com/featureflags/feature-api/internal/models"
)

func evalUserListRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	userIDs := toStringSlice(rule.Config["userIds"])
	if len(userIDs) == 0 {
		return false, false
	}

	for _, userID := range userIDs {
		if userID == ctx.UserID {
			return true, rule.Value
		}
	}
	return false, false
}
