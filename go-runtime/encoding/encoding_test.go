package encoding_test

import (
	"testing"

	. "github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/alecthomas/assert/v2"
)

func TestMarshal(t *testing.T) {
	type inner struct {
		FooBar string
	}
	somePtr := ftl.Some(42)
	tests := []struct {
		name     string
		input    any
		expected string
		err      string
	}{
		{name: "FieldRenaming", input: struct{ FooBar string }{""}, expected: `{"fooBar":""}`},
		{name: "String", input: struct{ String string }{"foo"}, expected: `{"string":"foo"}`},
		{name: "Int", input: struct{ Int int }{42}, expected: `{"int":42}`},
		{name: "Float", input: struct{ Float float64 }{42.42}, expected: `{"float":42.42}`},
		{name: "Bool", input: struct{ Bool bool }{true}, expected: `{"bool":true}`},
		{name: "Nil", input: struct{ Nil *int }{nil}, expected: `{"nil":null}`},
		{name: "Slice", input: struct{ Slice []int }{[]int{1, 2, 3}}, expected: `{"slice":[1,2,3]}`},
		{name: "SliceOfStrings", input: struct{ Slice []string }{[]string{"hello", "world"}}, expected: `{"slice":["hello","world"]}`},
		{name: "Map", input: struct{ Map map[string]int }{map[string]int{"foo": 42}}, expected: `{"map":{"foo":42}}`},
		{name: "Option", input: struct{ Option ftl.Option[int] }{ftl.Some(42)}, expected: `{"option":42}`},
		{name: "OptionPtr", input: struct{ Option *ftl.Option[int] }{&somePtr}, expected: `{"option":42}`},
		{name: "OptionStruct", input: struct{ Option ftl.Option[inner] }{ftl.Some(inner{"foo"})}, expected: `{"option":{"fooBar":"foo"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Marshal(tt.input)
			assert.EqualError(t, err, tt.err)
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}
