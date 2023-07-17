package compile

import (
	"testing"

	"github.com/alecthomas/assert/v2"
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
