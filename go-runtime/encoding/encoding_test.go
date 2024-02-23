package encoding_test

import (
	"reflect"
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
		{name: "Unit", input: ftl.Unit{}, expected: `{}`},
		{name: "UnitField", input: struct {
			String string
			Unit   ftl.Unit
		}{String: "something", Unit: ftl.Unit{}}, expected: `{"string":"something"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Marshal(tt.input)
			assert.EqualError(t, err, tt.err)
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type inner struct {
		FooBar string
	}
	somePtr := ftl.Some(42)
	tests := []struct {
		name     string
		input    string
		expected any
		err      string
	}{
		{name: "FieldRenaming", input: `{"fooBar":""}`, expected: struct{ FooBar string }{""}},
		{name: "String", input: `{"string":"foo"}`, expected: struct{ String string }{"foo"}},
		{name: "Int", input: `{"int":42}`, expected: struct{ Int int }{42}},
		{name: "Float", input: `{"float":42.42}`, expected: struct{ Float float64 }{42.42}},
		{name: "Bool", input: `{"bool":true}`, expected: struct{ Bool bool }{true}},
		{name: "Nil", input: `{"nil":null}`, expected: struct{ Nil *int }{nil}},
		{name: "Slice", input: `{"slice":[1,2,3]}`, expected: struct{ Slice []int }{[]int{1, 2, 3}}},
		{name: "SliceOfStrings", input: `{"slice":["hello","world"]}`, expected: struct{ Slice []string }{[]string{"hello", "world"}}},
		{name: "Map", input: `{"map":{"foo":42}}`, expected: struct{ Map map[string]int }{map[string]int{"foo": 42}}},
		{name: "Option", input: `{"option":42}`, expected: struct{ Option ftl.Option[int] }{ftl.Some(42)}},
		{name: "OptionPtr", input: `{"option":42}`, expected: struct{ Option *ftl.Option[int] }{&somePtr}},
		{name: "OptionStruct", input: `{"option":{"fooBar":"foo"}}`, expected: struct{ Option ftl.Option[inner] }{ftl.Some(inner{"foo"})}},
		{name: "Unit", input: `{}`, expected: ftl.Unit{}},
		{name: "UnitField", input: `{"string":"something"}`, expected: struct {
			String string
			Unit   ftl.Unit
		}{String: "something", Unit: ftl.Unit{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eType := reflect.TypeOf(tt.expected)
			if eType.Kind() == reflect.Ptr {
				eType = eType.Elem()
			}
			o := reflect.New(eType)
			err := Unmarshal([]byte(tt.input), o.Interface())
			assert.EqualError(t, err, tt.err)
			assert.Equal(t, tt.expected, o.Elem().Interface())
		})
	}
}

func TestRoundTrip(t *testing.T) {
	type inner struct {
		FooBar string
	}
	somePtr := ftl.Some(42)
	tests := []struct {
		name  string
		input any
	}{
		{name: "FieldRenaming", input: struct{ FooBar string }{""}},
		{name: "String", input: struct{ String string }{"foo"}},
		{name: "Int", input: struct{ Int int }{42}},
		{name: "Float", input: struct{ Float float64 }{42.42}},
		{name: "Bool", input: struct{ Bool bool }{true}},
		{name: "Nil", input: struct{ Nil *int }{nil}},
		{name: "Slice", input: struct{ Slice []int }{[]int{1, 2, 3}}},
		{name: "SliceOfStrings", input: struct{ Slice []string }{[]string{"hello", "world"}}},
		{name: "Map", input: struct{ Map map[string]int }{map[string]int{"foo": 42}}},
		{name: "Option", input: struct{ Option ftl.Option[int] }{ftl.Some(42)}},
		{name: "OptionPtr", input: struct{ Option *ftl.Option[int] }{&somePtr}},
		{name: "OptionStruct", input: struct{ Option ftl.Option[inner] }{ftl.Some(inner{"foo"})}},
		{name: "Unit", input: ftl.Unit{}},
		{name: "UnitField", input: struct {
			String string
			Unit   ftl.Unit
		}{String: "something", Unit: ftl.Unit{}}},
		{name: "Aliased", input: struct {
			TokenID string `json:"token_id"`
		}{"123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := Marshal(tt.input)
			assert.NoError(t, err)

			eType := reflect.TypeOf(tt.input)
			if eType.Kind() == reflect.Ptr {
				eType = eType.Elem()
			}
			o := reflect.New(eType)
			err = Unmarshal(marshaled, o.Interface())
			assert.NoError(t, err)

			assert.Equal(t, tt.input, o.Elem().Interface())
		})
	}
}
