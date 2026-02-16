// Package env provides utilities for parsing environment variables into nested Go data structures.
// It includes functions to filter unsafe variables, normalize keys, and construct hierarchical
// maps from environment variable names using specified separators.
package env

import (
	"strings"
)

const (
	objSep = ":"
	envSep = "_"
)

// ParseVariables processes a map of environment variables into a nested map structure.
// It takes the following parameters:
// - vars: map of environment variable key-value pairs to process
// - pre: prefix to filter variables by (variables not matching prefix are excluded)
// - sep: separator used in environment variable names to create nested structure
// - normalizeKey: whether to normalize keys by removing underscore separators
// Returns a nested map[string]any containing the processed environment variables.
func ParseVariables(vars map[string]string, pre, sep string, normalizeKey bool) map[string]any {
	data := make(map[string]any)

	pre = strings.ToLower(strings.TrimSpace(pre))

	for key, value := range vars {
		key = strings.ToLower(strings.TrimSpace(key))

		// Filter out unsafe variables
		if IsUnsafeVar(key) {
			continue
		}

		normalizedKey := key

		if pre != "" {
			if after, ok := strings.CutPrefix(key, pre); ok {
				normalizedKey = after
			} else {
				continue // Skip if doesn't match prefix
			}
		}

		// Save the prefix-stripped key before further modifications,
		// so we can also build nested maps using standard "_" separator.
		strippedKey := normalizedKey

		// The user's provided separator is usually "__", normalizing the keys
		// by removing '_' will likely break the user's provided separator.
		normalizedKey = strings.ReplaceAll(normalizedKey, sep, objSep)

		if normalizeKey {
			// Convert "snake_case_key" to "snakecasekey", this can be accessed later
			// as "snakeCaseKey" or "SnakeCaseKey".
			normalizedKey = strings.ReplaceAll(normalizedKey, envSep, "")

			if sep != "" {
				// Build nested map structure
				BuildNestedMap(data, normalizedKey, value, objSep)
			} else {
				data[normalizedKey] = value
			}
		}

		// Continue using the original keys anyway.
		if sep != "" {
			// Build nested map structure
			BuildNestedMap(data, key, value, sep)
		} else {
			data[key] = value
		}

		// Also build nested map using standard "_" separator, so that env vars
		// like "MY_ENV" are discoverable as nested keys (my -> env) in addition
		// to the longer namespace-based separator (e.g., "MY__ENV").
		if sep != "" && sep != envSep {
			BuildNestedMap(data, strippedKey, value, envSep)
		}
	}

	return data
}
