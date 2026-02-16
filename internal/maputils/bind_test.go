package maputils_test

import (
	"testing"
	"time"

	"github.com/gasmod/gas-config/internal/maputils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBind_EmbeddedStruct(t *testing.T) {
	t.Parallel()

	type Embedded struct {
		MyKey string `json:"mykey"`
		Value int    `json:"value"`
	}

	type Parent struct {
		Name string `json:"name"`
		Embedded
	}

	src := map[string]any{
		"mykey": "embedded_value",
		"value": 42,
		"name":  "parent_name",
	}

	var dest Parent
	require.NoError(t, maputils.Bind(src, &dest))

	// These should work with embedded structs
	assert.Equal(t, "embedded_value", dest.MyKey)
	assert.Equal(t, 42, dest.Value)
	assert.Equal(t, "parent_name", dest.Name)
}

func TestBind_NestedEmbeddedStruct(t *testing.T) {
	t.Parallel()

	type Level1 struct {
		Field1 string `json:"field1"`
	}

	type Level2 struct {
		Level1

		Field2 string `json:"field2"`
	}

	type Parent struct {
		Level2

		ParentField string `json:"parent_field"`
	}

	src := map[string]any{
		"field1":       "level1_value",
		"field2":       "level2_value",
		"parent_field": "parent_value",
	}

	var dest Parent
	require.NoError(t, maputils.Bind(src, &dest))

	assert.Equal(t, "level1_value", dest.Field1)
	assert.Equal(t, "level2_value", dest.Field2)
	assert.Equal(t, "parent_value", dest.ParentField)
}

func TestBind_EmbeddedStructWithPointer(t *testing.T) {
	t.Parallel()

	type Embedded struct {
		MyKey string `json:"mykey"`
	}

	type Parent struct {
		*Embedded

		Name string `json:"name"`
	}

	src := map[string]any{
		"mykey": "embedded_value",
		"name":  "parent_name",
	}

	var dest Parent
	require.NoError(t, maputils.Bind(src, &dest))

	assert.Equal(t, "parent_name", dest.Name)
	// Embedded pointer should be initialized and populated
	require.NotNil(t, dest.Embedded)
	assert.Equal(t, "embedded_value", dest.MyKey)
}

func TestUnbind_EmbeddedStruct(t *testing.T) {
	t.Parallel()

	type Embedded struct {
		MyKey string `json:"mykey"`
		Value int    `json:"value"`
	}

	type Parent struct {
		Name string `json:"name"`
		Embedded
	}

	src := Parent{
		Embedded: Embedded{
			MyKey: "embedded_value",
			Value: 42,
		},
		Name: "parent_name",
	}

	dest := make(map[string]any)
	require.NoError(t, maputils.Unbind(&src, dest))

	// All fields should be available at the top level
	assert.Equal(t, "embedded_value", dest["mykey"])
	assert.Equal(t, 42, dest["value"])
	assert.Equal(t, "parent_name", dest["name"])
}

func TestBind_DurationField(t *testing.T) {
	t.Parallel()

	type Struct struct {
		Int64    time.Duration
		Duration time.Duration
		String0  time.Duration
		String1  time.Duration
		String2  time.Duration
	}

	src := map[string]any{
		"int64":    int64(1 * time.Second),
		"duration": 2 * time.Minute,
		"string0":  "300ms",
		"string1":  "-1.5h",
		"string2":  "2h45m",
	}

	var dest Struct
	require.NoError(t, maputils.Bind(src, &dest))

	assert.Equal(t, 1*time.Second, dest.Int64)
	assert.Equal(t, 2*time.Minute, dest.Duration)
	assert.Equal(t, 300*time.Millisecond, dest.String0)
	assert.Equal(t, -(90 * time.Minute), dest.String1)
	assert.Equal(t, 165*time.Minute, dest.String2)
}

func TestBind_Array(t *testing.T) {
	t.Parallel()

	type Struct struct {
		Str []string
		Int []int
	}

	src := map[string]any{
		"str": "a,b,c",
		"int": "1,2,3",
	}

	var dest Struct
	require.NoError(t, maputils.Bind(src, &dest))

	assert.Equal(t, []string{"a", "b", "c"}, dest.Str)
	assert.Equal(t, []int{1, 2, 3}, dest.Int)
}
