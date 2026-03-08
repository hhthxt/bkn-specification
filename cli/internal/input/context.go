package input

import (
	"fmt"
	"strings"
)

func ParseKeyValuePairs(values []string) (map[string]any, error) {
	out := make(map[string]any, len(values))
	for _, value := range values {
		key, raw, ok := strings.Cut(value, "=")
		key = strings.TrimSpace(key)
		raw = strings.TrimSpace(raw)
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid key=value pair %q", value)
		}
		out[key] = raw
	}
	return out, nil
}
