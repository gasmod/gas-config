package maputils

import "strings"

// LowercaseKeys recursively converts all keys in the map to lowercase, in place.
func LowercaseKeys(m map[string]any) {
	for k, v := range m {
		normalK := strings.ToLower(strings.TrimSpace(k))
		if normalK != k {
			delete(m, k)
			m[normalK] = v
		}

		// Recurse if value is a nested map
		if sv, ok := v.(map[string]any); ok {
			LowercaseKeys(sv)
		}
	}
}
