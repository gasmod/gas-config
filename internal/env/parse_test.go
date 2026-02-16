package env_test

import (
	"testing"

	"github.com/ahmedkamalio/gcfg/internal/env"
	"github.com/stretchr/testify/assert"
)

func TestParseVariables_StandardNameDiscovery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		vars         map[string]string
		pre          string
		sep          string
		normalizeKey bool
		check        func(t *testing.T, data map[string]any)
	}{
		{
			name:         "single underscore creates nested map",
			vars:         map[string]string{"MY_ENV": "val"},
			sep:          "__",
			normalizeKey: true,
			check: func(t *testing.T, data map[string]any) {
				t.Helper()

				// Nested via standard "_" separator
				nested, ok := data["my"].(map[string]any)
				assert.True(t, ok, "expected nested map under 'my'")
				assert.Equal(t, "val", nested["env"])

				// Original flat key still accessible
				assert.Equal(t, "val", data["my_env"])

				// Normalized key still accessible
				assert.Equal(t, "val", data["myenv"])
			},
		},
		{
			name:         "double underscore still works for nesting",
			vars:         map[string]string{"MY__ENV": "val"},
			sep:          "__",
			normalizeKey: true,
			check: func(t *testing.T, data map[string]any) {
				t.Helper()

				nested, ok := data["my"].(map[string]any)
				assert.True(t, ok, "expected nested map under 'my'")
				assert.Equal(t, "val", nested["env"])
			},
		},
		{
			name:         "both single and double underscore vars coexist",
			vars:         map[string]string{"DB_HOST": "single", "DB__PORT": "double"},
			sep:          "__",
			normalizeKey: true,
			check: func(t *testing.T, data map[string]any) {
				t.Helper()

				nested, ok := data["db"].(map[string]any)
				assert.True(t, ok, "expected nested map under 'db'")
				assert.Equal(t, "single", nested["host"])
				assert.Equal(t, "double", nested["port"])
			},
		},
		{
			name:         "multi-level nesting with single underscore",
			vars:         map[string]string{"DB_HOST_PORT": "5432"},
			sep:          "__",
			normalizeKey: true,
			check: func(t *testing.T, data map[string]any) {
				t.Helper()

				db, ok := data["db"].(map[string]any)
				assert.True(t, ok, "expected nested map under 'db'")
				host, ok := db["host"].(map[string]any)
				assert.True(t, ok, "expected nested map under 'host'")
				assert.Equal(t, "5432", host["port"])
			},
		},
		{
			name:         "with prefix strips prefix before nesting",
			vars:         map[string]string{"APP_DB_HOST": "localhost"},
			pre:          "APP_",
			sep:          "__",
			normalizeKey: true,
			check: func(t *testing.T, data map[string]any) {
				t.Helper()

				nested, ok := data["db"].(map[string]any)
				assert.True(t, ok, "expected nested map under 'db'")
				assert.Equal(t, "localhost", nested["host"])
			},
		},
		{
			name:         "no standard name nesting when sep is single underscore",
			vars:         map[string]string{"MY_ENV": "val"},
			sep:          "_",
			normalizeKey: true,
			check: func(t *testing.T, data map[string]any) {
				t.Helper()

				// When sep is already "_", the existing nesting handles it
				nested, ok := data["my"].(map[string]any)
				assert.True(t, ok, "expected nested map under 'my'")
				assert.Equal(t, "val", nested["env"])
			},
		},
		{
			name:         "no standard name nesting when sep is empty",
			vars:         map[string]string{"MY_ENV": "val"},
			sep:          "",
			normalizeKey: true,
			check: func(t *testing.T, data map[string]any) {
				t.Helper()

				// No nesting when sep is empty, original flat key is accessible
				assert.Equal(t, "val", data["my_env"])
				// No nested structure created
				assert.Nil(t, data["my"])
			},
		},
		{
			name:         "standard name nesting without normalize",
			vars:         map[string]string{"MY_ENV": "val"},
			sep:          "__",
			normalizeKey: false,
			check: func(t *testing.T, data map[string]any) {
				t.Helper()

				// Standard "_" nesting still works without normalization
				nested, ok := data["my"].(map[string]any)
				assert.True(t, ok, "expected nested map under 'my'")
				assert.Equal(t, "val", nested["env"])

				// Original flat key still accessible
				assert.Equal(t, "val", data["my_env"])

				// Normalized key should NOT exist
				assert.Nil(t, data["myenv"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := env.ParseVariables(tt.vars, tt.pre, tt.sep, tt.normalizeKey)
			tt.check(t, data)
		})
	}
}
