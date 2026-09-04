package bootutil

import (
	"fmt"
	"strconv"
	"strings"
)

func AsString(value any) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprint(value)
}

func CoerceStringMap(v any) map[string]string {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]string); ok {
		return m
	}
	if m, ok := v.(map[string]any); ok {
		out := make(map[string]string, len(m))
		for k, val := range m {
			out[k] = AsString(val)
		}
		return out
	}
	return nil
}

func AsInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		return n, err == nil
	default:
		n, err := strconv.Atoi(strings.TrimSpace(fmt.Sprint(value)))
		return n, err == nil
	}
}
