package compile

import (
	"go/ast"
	"go/types"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/TBD54566975/ftl/backend/schema"
)

func TestExtractModuleSchema(t *testing.T) {
	actual, err := ExtractModuleSchema("testdata/one")
	assert.NoError(t, err)
	actual = schema.Normalise(actual)
	expected := `module one {
  data Nested {
  }

  data Req {
    int Int
    int64 Int
    float Float
    string String
    slice [String]
    map {String: String}
    nested one.Nested
    optional one.Nested?
    time Time
    user two.User
    bytes Bytes
  }

  data Resp {
  }

  verb verb(one.Req) one.Resp
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
		{name: "Module", input: "ftl:module foo", expected: &directiveModule{Name: "foo"}},
		{name: "Verb", input: "ftl:verb", expected: &directiveVerb{Verb: true}},
		{name: "IngressImplicitFTL", input: `ftl:ingress GET /foo`, expected: &directiveIngress{
			Type: &directiveIngressHTTP{
				Method: "GET",
				Path:   "/foo",
			},
		}},
		{name: "IngressFTL", input: `ftl:ingress ftl POST /bar`, expected: &directiveIngress{
			Type: &directiveIngressHTTP{
				Type:   "ftl",
				Method: "POST",
				Path:   "/bar",
			},
		}},
		{name: "IngressHTTP", input: `ftl:ingress http POST /bar`, expected: &directiveIngress{
			Type: &directiveIngressHTTP{
				Type:   "http",
				Method: "POST",
				Path:   "/bar",
			},
		}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := directiveParser.ParseString("", tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got.Directive, assert.Exclude[lexer.Position]())
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
