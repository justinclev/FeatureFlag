package evaluator

import (
	"strings"
	"sync"

	"github.com/featureflags/feature-api/internal/models"
)

var compiledUserLists sync.Map // map[string]map[string]struct{} (Key: Rule ID)

func evalUserListRule(rule models.Rule, ctx models.EvaluationContext) (bool, bool) {
	rawIDs := getConfig(rule.Config, "userIds")
	userIDs := toStringSlice(rawIDs)

	if len(userIDs) == 0 {
		return false, false
	}

	actualUserID := strings.ToLower(strings.TrimSpace(ctx.UserID))

	// Principal optimization: Compile large lists into a map for O(1) lookup
	// and cache them per-rule across evaluations to save GC pressure.
	if len(userIDs) > 20 {
		ruleID := rule.ID.Hex()
		
		// Fast path: load compiled map
		if cachedMap, ok := compiledUserLists.Load(ruleID); ok {
			fastMap := cachedMap.(map[string]struct{})
			if _, matched := fastMap[actualUserID]; matched {
				return true, rule.Value
			}
			return false, false
		}

		// Slow path: compile and store
		fastMap := make(map[string]struct{}, len(userIDs))
		for _, id := range userIDs {
			fastMap[strings.ToLower(strings.TrimSpace(id))] = struct{}{}
		}
		compiledUserLists.Store(ruleID, fastMap)

		if _, matched := fastMap[actualUserID]; matched {
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
