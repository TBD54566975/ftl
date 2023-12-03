package compile

import (
	"go/ast"
	"go/types"
	"testing"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/alecthomas/assert/v2"
)

func TestExtractModuleSchema(t *testing.T) {
	actual, err := ExtractModuleSchema("testdata")
	assert.NoError(t, err)
	actual = schema.Normalise(actual)
	expected := `module main {
  data Nested {
  }

  data Req {
    int Int
    int64 Int
    float Float
    string String
    slice [String]
    map {String: String}
    nested Nested
    optional Nested?
    time Time
  }

  data Resp {
  }

  verb verb(Req) Resp
}
`
	assert.Equal(t, expected, actual.String())
}

func TestParseDirectives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected directive
	}{
		{name: "NoAttributes", input: "ftl:verb", expected: directive{Kind: "verb"}},
		{name: "PositionalAttribute", input: `ftl:ingress GET /foo`, expected: directive{
			Kind: "ingress",
			Attrs: []directiveAttr{
				dirAttrIdent("", "GET"),
				dirAttrPath("", "/foo"),
			},
		}},
		{name: "MixedPositionalKeywordAttributes", input: `ftl:ingress POST path=/bar`,
			expected: directive{
				Kind: "ingress",
				Attrs: []directiveAttr{
					dirAttrIdent("", "POST"),
					dirAttrPath("path", "/bar"),
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := directiveParser.ParseString("", tt.input)
			assert.NoError(t, err)
			assert.Equal(t, &tt.expected, got)
		})
	}
}

func TestParseTypesTime(t *testing.T) {
	timeRef := mustLoadRef("time", "Time").Type()
	parsed, err := parseType(nil, &ast.Ident{}, timeRef)
	assert.NoError(t, err)
	_, ok := parsed.(*schema.Time)
	assert.True(t, ok)
}

func TestParseBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Type
		expected schema.Type
	}{
		{name: "String", input: types.Typ[types.String], expected: &schema.String{}},
		{name: "Int", input: types.Typ[types.Int], expected: &schema.Int{}},
		{name: "Bool", input: types.Typ[types.Bool], expected: &schema.Bool{}},
		{name: "Float64", input: types.Typ[types.Float64], expected: &schema.Float{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseType(nil, &ast.Ident{}, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}

func dirAttrIdent(key string, s string) directiveAttr {
	if key == "" {
		return directiveAttr{Value: directiveValue{Ident: &s}}
	}
	return directiveAttr{Key: &key, Value: directiveValue{Ident: &s}}
}
func dirAttrPath(key string, s string) directiveAttr {
	if key == "" {
		return directiveAttr{Value: directiveValue{Path: &s}}
	}
	return directiveAttr{Key: &key, Value: directiveValue{Path: &s}}
}
