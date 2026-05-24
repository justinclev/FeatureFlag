package evaluator

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func toString(v any) string {
	if v == nil {
		return ""
	}
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

func toSafeTime(v any) (time.Time, bool) {
	if v == nil {
		return time.Time{}, false
	}
	switch v := v.(type) {
	case time.Time:
		return v.UTC(), true
	case bson.DateTime:
		return v.Time().UTC(), true
	case int64:
		return time.UnixMilli(v).UTC(), true
	case string:
		if v == "" {
			return time.Time{}, false
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.UTC(), true
		}
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return t.UTC(), true
		}
		if t, err := time.Parse("2006-01-02T15:04:05Z", v); err == nil {
			return t.UTC(), true
		}
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}

	// Use reflection for maximum resiliency to any slice/array type (primitive.A, []any, etc)
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		res := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			res[i] = toString(rv.Index(i).Interface())
		}
		return res
	}

	return []string{toString(v)}
}

func getConfig(m map[string]any, key string) any {
	if m == nil {
		return nil
	}
	if v, ok := m[key]; ok {
		return v
	}
	for k, v := range m {
		if stringsEqualFold(k, key) {
			return v
		}
	}
	return nil
}

func stringsEqualFold(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		c1, c2 := s1[i], s2[i]
		if c1 != c2 {
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				return false
			}
		}
	}
	return true
}
