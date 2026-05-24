package evaluator

import (
	"strings"

	"github.com/featureflags/feature-api/internal/models"
)

func evalUserListRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	userIDs := toStringSlice(rule.Config["userIds"])
	if len(userIDs) == 0 {
		return false, false
	}

	// Principal optimization: Case-insensitive and trimmed comparison for resilient matching.
	actualUserID := strings.ToLower(strings.TrimSpace(ctx.UserID))

	for _, userID := range userIDs {
		if strings.ToLower(strings.TrimSpace(userID)) == actualUserID {
			return true, rule.Value
		}
	}
	return false, false
}
