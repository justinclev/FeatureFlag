package evaluator

import (
	"github.com/featureflags/feature-api/internal/models"
)

func evalUserListRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	for _, userID := range rule.Config.UserIDs {
		if userID == ctx.UserID {
			return true, rule.Value
		}
	}
	return false, false
}
