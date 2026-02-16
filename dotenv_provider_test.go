package config_test

import (
	"os"
	"testing"
	"testing/fstest"

	config "github.com/gasmod/gas-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDotEnvProvider_DefaultOptions(t *testing.T) {
	t.Parallel()

	p := config.NewDotEnvProvider()
	_, err := p.Load()
	require.Error(t, err)
}

func TestDotEnvProvider_WithDotEnvFile_FileNotFound(t *testing.T) {
	t.Parallel()

	p := config.NewDotEnvProvider(
		config.WithDotEnvFilePath(".env.non-existing"),
	)
	_, err := p.Load()
	require.Error(t, err)
}

func TestDotEnvProvider_WithDotEnvFile(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		".env": &fstest.MapFile{
			Data: []byte(`
				TEST_KEY=test_value
			`),
		},
	}

	p := config.NewDotEnvProvider(
		config.WithDotEnvFilePath(".env"),
		config.WithDotEnvFileFS(&fsys),
	)

	values, err := p.Load()
	require.NoError(t, err)
	assert.Equal(t, "test_value", values["testkey"])
}

func TestDotEnvProvider_WithEnvSeparator(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		".env": &fstest.MapFile{
			Data: []byte(`
				TEST__KEY=test_value
			`),
		},
	}

	p := config.NewDotEnvProvider(
		config.WithDotEnvFilePath(".env"),
		config.WithDotEnvFileFS(&fsys),
		config.WithDotEnvSeparator("__"),
	)

	values, err := p.Load()
	require.NoError(t, err)
	assert.IsType(t, map[string]any{}, values["test"])
	assert.Equal(t, "test_value", values["test"].(map[string]any)["key"])
}

func TestDotEnvProvider_WithEnvNormalizeVarNames(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		".env": &fstest.MapFile{
			Data: []byte(`
				TEST_KEY=test_value
			`),
		},
	}

	p := config.NewDotEnvProvider(
		config.WithDotEnvFilePath(".env"),
		config.WithDotEnvFileFS(&fsys),
		config.WithDotEnvNormalizeVarNames(false), // keep original variable names
	)

	values, err := p.Load()
	require.NoError(t, err)
	assert.Equal(t, "test_value", values["test_key"])
}

func TestDotEnvProvider_Syntax(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		".env": &fstest.MapFile{
			Data: []byte(`
				# This is a comment
				TEST_KEY=test_value
				TEST_KEY2=test_value2 # This is an inline comment
			`),
		},
	}

	p := config.NewDotEnvProvider(
		config.WithDotEnvFilePath(".env"),
		config.WithDotEnvFileFS(&fsys),
	)

	values, err := p.Load()
	require.NoError(t, err)
	assert.Equal(t, "test_value", values["testkey"])
	assert.Equal(t, "test_value2", values["testkey2"])
}

func TestDotEnvProvider_WithDotEnvFile_FileNotFoundNoPanic(t *testing.T) {
	t.Parallel()

	p := config.NewDotEnvProvider(
		config.WithDotEnvFilePath(".env.non-existing"),
		config.WithDotEnvFileNotFoundPanic(false),
	)
	_, err := p.Load()
	require.NoError(t, err)
}

func TestDotEnvProvider_AppendToOSEnv(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		".env": &fstest.MapFile{
			Data: []byte("TEST_KEY=test_value"),
		},
	}

	p := config.NewDotEnvProvider(
		config.WithDotEnvFilePath(".env"),
		config.WithDotEnvFileFS(&fsys),
	)

	values, err := p.Load()
	require.NoError(t, err)
	assert.Equal(t, "test_value", values["testkey"])
	assert.Equal(t, "test_value", os.Getenv("TEST_KEY"))
}

func TestDotEnvProvider_NoAppendToOSEnv(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		".env": &fstest.MapFile{
			Data: []byte("MY_KEY=test_value"),
		},
	}

	p := config.NewDotEnvProvider(
		config.WithDotEnvFilePath(".env"),
		config.WithDotEnvFileFS(&fsys),
		config.WithDotEnvFileAppendToOSEnv(false),
	)

	values, err := p.Load()
	require.NoError(t, err)
	assert.Equal(t, "test_value", values["mykey"])
	assert.Empty(t, os.Getenv("MY_KEY"), "Expected os.Getenv(\"MY_KEY\") to be empty")
}
