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
		// JSON numbers unmarshal as float64
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
