package ingress

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
)

func TestTransformFromAliasedFields(t *testing.T) {
	schemaText := `
		module test {
			data Inner {
				waz String alias foo
			}

			data Test {
				scalar String alias bar
				inner Inner
				array [Inner]
				map {String: Inner}
				optional Inner
			}
		}
		`
	sch, err := schema.ParseString("test", schemaText)
	assert.NoError(t, err)
	actual, err := transformFromAliasedFields(&schema.DataRef{Module: "test", Name: "Test"}, sch, map[string]any{
		"bar": "value",
		"inner": map[string]any{
			"foo": "value",
		},
		"array": []any{
			map[string]any{
				"foo": "value",
			},
		},
		"map": map[string]any{
			"key": map[string]any{
				"foo": "value",
			},
		},
		"optional": map[string]any{
			"foo": "value",
		},
	})
	expected := map[string]any{
		"scalar": "value",
		"inner": map[string]any{
			"waz": "value",
		},
		"array": []any{
			map[string]any{
				"waz": "value",
			},
		},
		"map": map[string]any{
			"key": map[string]any{
				"waz": "value",
			},
		},
		"optional": map[string]any{
			"waz": "value",
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestTransformToAliasedFields(t *testing.T) {
	schemaText := `
		module test {
			data Inner {
				waz String alias foo
			}

			data Test {
				scalar String alias bar
				inner Inner
				array [Inner]
				map {String: Inner}
				optional Inner
			}
		}
		`
	sch, err := schema.ParseString("test", schemaText)
	assert.NoError(t, err)
	actual, err := transformToAliasedFields(&schema.DataRef{Module: "test", Name: "Test"}, sch, map[string]any{
		"scalar": "value",
		"inner": map[string]any{
			"waz": "value",
		},
		"array": []any{
			map[string]any{
				"waz": "value",
			},
		},
		"map": map[string]any{
			"key": map[string]any{
				"waz": "value",
			},
		},
		"optional": map[string]any{
			"waz": "value",
		},
	})
	expected := map[string]any{
		"bar": "value",
		"inner": map[string]any{
			"foo": "value",
		},
		"array": []any{
			map[string]any{
				"foo": "value",
			},
		},
		"map": map[string]any{
			"key": map[string]any{
				"foo": "value",
			},
		},
		"optional": map[string]any{
			"foo": "value",
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
