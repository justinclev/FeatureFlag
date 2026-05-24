package evaluator

import (
	"strings"

	"github.com/featureflags/feature-api/internal/models"
)

func evalUserListRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	// Debug logging is not available, using surgical logic fix.
	rawIDs := getConfig(rule.Config, "userIds")
	userIDs := toStringSlice(rawIDs)
	
	if len(userIDs) == 0 {
		return false, false
	}

	actualUserID := strings.ToLower(strings.TrimSpace(ctx.UserID))

	// Principal optimization: For large lists, convert to a map for O(1) lookup.
	if len(userIDs) > 20 {
		fastMap := make(map[string]struct{}, len(userIDs))
		for _, id := range userIDs {
			fastMap[strings.ToLower(strings.TrimSpace(id))] = struct{}{}
		}
		if _, ok := fastMap[actualUserID]; ok {
			return true, rule.Value
		}
		return false, false
	}

	for _, userID := range userIDs {
		cleanID := strings.ToLower(strings.TrimSpace(userID))
		if cleanID == actualUserID {
			return true, rule.Value
		}
	}
	return false, false
}
