package gcfg_test

import (
	"testing"

	"github.com/ahmedkamalio/gcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvProvider_DefaultOptions(t *testing.T) {
	t.Setenv("TEST_KEY", "test_value")

	p := gcfg.NewEnvProvider()

	values, err := p.Load()
	require.NoError(t, err)

	// Value can be accessed using both original AND normalized names
	assert.Equal(t, "test_value", values["test_key"])
	assert.Equal(t, "test_value", values["testkey"])
}

func TestEnvProvider_WithEnvPrefix(t *testing.T) {
	t.Setenv("TEST_KEY", "unaccessible_value")
	t.Setenv("MYAPP_TEST_KEY", "test_value")

	p := gcfg.NewEnvProvider(
		gcfg.WithEnvPrefix("MYAPP_"), // load only prefixed variables
	)

	values, err := p.Load()
	require.NoError(t, err)
	assert.Equal(t, "test_value", values["testkey"])
}

func TestEnvProvider_WithEnvSeparator(t *testing.T) {
	t.Setenv("TEST__KEY", "test_value")

	p := gcfg.NewEnvProvider(
		gcfg.WithEnvSeparator("__"),
	)

	values, err := p.Load()
	require.NoError(t, err)
	assert.IsType(t, map[string]any{}, values["test"])
	assert.Equal(t, "test_value", values["test"].(map[string]any)["key"])
}

func TestEnvProvider_StandardNameDiscovery(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")

	p := gcfg.NewEnvProvider()

	values, err := p.Load()
	require.NoError(t, err)

	t.Log(values)

	// Value can be accessed as nested map via standard "_" separator
	nested, ok := values["db"].(map[string]any)
	assert.True(t, ok, "expected nested map under 'db'")
	assert.Equal(t, "localhost", nested["host"])

	// Original flat key still accessible
	assert.Equal(t, "localhost", values["db_host"])

	// Normalized key still accessible
	assert.Equal(t, "localhost", values["dbhost"])
}

func TestEnvProvider_StandardNameDiscoveryWithPrefix(t *testing.T) {
	t.Setenv("MYAPP_DB_HOST", "localhost")

	p := gcfg.NewEnvProvider(
		gcfg.WithEnvPrefix("MYAPP_"),
	)

	values, err := p.Load()
	require.NoError(t, err)

	t.Log(values)

	// After prefix stripping, "DB_HOST" is nested via "_"
	nested, ok := values["db"].(map[string]any)
	assert.True(t, ok, "expected nested map under 'db'")
	assert.Equal(t, "localhost", nested["host"])

	// Original flat key still accessible
	assert.Equal(t, "localhost", values["myapp_db_host"])

	// Normalized key still accessible
	assert.Equal(t, "localhost", values["dbhost"])
}

func TestEnvProvider_WithEnvNormalizeVarNames(t *testing.T) {
	t.Setenv("TEST_KEY", "test_value")

	p := gcfg.NewEnvProvider(
		gcfg.WithEnvNormalizeVarNames(false),
	)

	values, err := p.Load()
	require.NoError(t, err)

	// Value can only be accessed using the original names
	assert.Equal(t, "test_value", values["test_key"])
	assert.Empty(t, values["testkey"])
}
