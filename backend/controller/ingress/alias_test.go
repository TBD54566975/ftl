package ingress

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
)

func TestTransformFromAliasedFields(t *testing.T) {
	schemaText := `
		module test {
			enum TypeEnum {
				A test.Inner
				B String
			}
			
			data Inner {
				waz String +alias json "foo"
			}

			data Test {
				scalar String +alias json "bar"
				inner test.Inner
				array [test.Inner]
				map {String: test.Inner}
				optional test.Inner
				typeEnum test.TypeEnum
			}
		}
		`

	sch, err := schema.ParseString("test", schemaText)
	assert.NoError(t, err)
	actual, err := transformFromAliasedFields(&schema.Ref{Module: "test", Name: "Test"}, sch, map[string]any{
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
		"typeEnum": map[string]any{
			"name":  "A",
			"value": map[string]any{"foo": "value"},
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
		"typeEnum": map[string]any{
			"name":  "A",
			"value": map[string]any{"waz": "value"},
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestTransformToAliasedFields(t *testing.T) {
	schemaText := `
		module test {
			enum TypeEnum {
				A test.Inner
				B String
			}

			data Inner {
				waz String +alias json "foo"
			}

			data Test {
				scalar String +alias json "bar"
				inner test.Inner
				array [test.Inner]
				map {String: test.Inner}
				optional test.Inner
				typeEnum test.TypeEnum
			}
		}
		`

	sch, err := schema.ParseString("test", schemaText)
	assert.NoError(t, err)
	actual, err := transformToAliasedFields(&schema.Ref{Module: "test", Name: "Test"}, sch, map[string]any{
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
		"typeEnum": map[string]any{
			"name":  "A",
			"value": map[string]any{"waz": "value"},
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
		"typeEnum": map[string]any{
			"name":  "A",
			"value": map[string]any{"foo": "value"},
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
