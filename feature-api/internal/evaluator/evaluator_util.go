package evaluator

import (
	"fmt"
	"strconv"
)

func toString(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case bool:
		return strconv.FormatBool(v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toSafeFloat(v any) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch v := v.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	switch v := v.(type) {
	case []string:
		return v
	case []any:
		res := make([]string, len(v))
		for i, val := range v {
			res[i] = toString(val)
		}
		return res
	default:
		return nil
	}
}
