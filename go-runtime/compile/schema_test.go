package compile

import (
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/alecthomas/assert/v2"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"reflect"
	"testing"
)

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
		t.Run(tt.name, func(t *testing.T) {
			got, err := directiveParser.ParseString("", tt.input)
			assert.NoError(t, err)
			assert.Equal(t, &tt.expected, got)
		})
	}
}

func TestParseTypesTime(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, "time")
	if err != nil {
		log.Fatal(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return
	}

	timePkg := pkgs[0]
	timeTime := timePkg.Types.Scope().Lookup("Time").Type()

	parsed, err := parseType(nil, &ast.Ident{}, timeTime)
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
			if !reflect.DeepEqual(reflect.TypeOf(parsed), reflect.TypeOf(tt.expected)) {
				t.Errorf("Type mismatch: got %T, want %T", parsed, tt.expected)
			}
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
