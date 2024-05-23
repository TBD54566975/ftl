package encoding_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	. "github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type discriminator interface {
	tag()
}

type unregistered interface {
	unregistered()
}

type variant struct {
	Message string
}

func (variant) tag()          {}
func (variant) unregistered() {}

func TestMarshal(t *testing.T) {
	type inner struct {
		FooBar string
	}
	type sumtypeStruct struct {
		D discriminator
	}
	type validateOmitempty struct {
		ShouldOmit   string `json:",omitempty"`
		ShouldntOmit string `json:""`
		NotTagged    string
	}
	type validateOmitemptyOption struct {
		ShouldOmit   ftl.Option[string] `json:",omitempty"`
		ShouldntOmit ftl.Option[string] `json:""`
		NotTagged    ftl.Option[string]
	}
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
		{name: "Slice", input: struct{ Slice []int }{[]int{1, 2, 3}}, expected: `{"slice":[1,2,3]}`},
		{name: "SliceOfStrings", input: struct{ Slice []string }{[]string{"hello", "world"}}, expected: `{"slice":["hello","world"]}`},
		{name: "Map", input: struct{ Map map[string]int }{map[string]int{"foo": 42}}, expected: `{"map":{"foo":42}}`},
		{name: "Option", input: struct{ Option ftl.Option[int] }{ftl.Some(42)}, expected: `{"option":42}`},
		{name: "OptionNull", input: struct{ Option ftl.Option[int] }{ftl.None[int]()}, expected: `{"option":null}`},
		{name: "OptionZero", input: struct{ Option ftl.Option[int] }{ftl.Some(0)}, expected: `{"option":0}`},
		{name: "OptionStruct", input: struct{ Option ftl.Option[inner] }{ftl.Some(inner{"foo"})}, expected: `{"option":{"fooBar":"foo"}}`},
		{name: "OptionSumType", input: struct{ Option ftl.Option[sumtypeStruct] }{ftl.Some(sumtypeStruct{variant{"hello"}})}, expected: `{"option":{"d":{"name":"Variant","value":{"message":"hello"}}}}`},
		{name: "Unit", input: ftl.Unit{}, expected: `{}`},
		{name: "UnitField", input: struct {
			String string
			Unit   ftl.Unit
		}{String: "something", Unit: ftl.Unit{}}, expected: `{"string":"something","unit":{}}`},
		{name: "Pointer", input: &struct{ String string }{"foo"}, err: `pointer types are not supported: *struct { String string }`},
		{name: "SumType", input: struct{ D discriminator }{variant{"hello"}}, expected: `{"d":{"name":"Variant","value":{"message":"hello"}}}`},
		{name: "UnregisteredSumType", input: struct{ D unregistered }{variant{"hello"}}, err: `the only supported interface types are enums or any, not encoding_test.unregistered`},
		{name: "OmitEmptyNotNull", input: validateOmitempty{"foo", "bar", "baz"}, expected: `{"shouldOmit":"foo","shouldntOmit":"bar","notTagged":"baz"}`},
		{name: "OmitEmptyNull", input: validateOmitempty{}, expected: `{"shouldntOmit":"","notTagged":""}`},
		{name: "OmitEmptyOptionNone", input: validateOmitemptyOption{
			ShouldOmit:   ftl.None[string](),
			ShouldntOmit: ftl.None[string](),
			NotTagged:    ftl.None[string](),
		}, expected: `{"shouldntOmit":null,"notTagged":null}`},
	}

	reflection.AllowAnyPackageForTesting = true
	defer func() { reflection.AllowAnyPackageForTesting = false }()
	reflection.ResetTypeRegistry()
	reflection.Register(reflection.WithSumType[discriminator](variant{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Marshal(tt.input)
			assert.EqualError(t, err, tt.err)
			if err == nil {
				assert.Equal(t, tt.expected, string(actual))
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type inner struct {
		FooBar string
	}
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
		{name: "Slice", input: `{"slice":[1,2,3]}`, expected: struct{ Slice []int }{[]int{1, 2, 3}}},
		{name: "SliceOfStrings", input: `{"slice":["hello","world"]}`, expected: struct{ Slice []string }{[]string{"hello", "world"}}},
		{name: "Map", input: `{"map":{"foo":42}}`, expected: struct{ Map map[string]int }{map[string]int{"foo": 42}}},
		{name: "OptionNull", input: `{"option":null}`, expected: struct{ Option ftl.Option[int] }{ftl.None[int]()}},
		{name: "OptionNullWhitespace", input: `{"option": null}`, expected: struct{ Option ftl.Option[int] }{ftl.None[int]()}},
		{name: "OptionZero", input: `{"option":0}`, expected: struct{ Option ftl.Option[int] }{ftl.Some(0)}},
		{name: "Option", input: `{"option":42}`, expected: struct{ Option ftl.Option[int] }{ftl.Some(42)}},
		{name: "OptionStruct", input: `{"option":{"fooBar":"foo"}}`, expected: struct{ Option ftl.Option[inner] }{ftl.Some(inner{"foo"})}},
		{name: "Unit", input: `{}`, expected: ftl.Unit{}},
		{name: "UnitField", input: `{"string":"something"}`, expected: struct {
			String string
			Unit   ftl.Unit
		}{String: "something", Unit: ftl.Unit{}}},
		// Whitespaces after each `:` and multiple fields to test handling of the
		// two potential terminal delimiters: `}` and `,`
		{name: "ComplexFormatting", input: `{"option": null, "bool": true}`, expected: struct {
			Option ftl.Option[int]
			Bool   bool
		}{ftl.None[int](), true}},
		{name: "Pointer", input: `{"string":"foo"}`, expected: &struct{ String string }{}, err: `pointer types are not supported: *struct { String string }`},
		{name: "SumType", input: `{"d":{"name":"Variant","value":{"message":"hello"}}}`, expected: struct{ D discriminator }{variant{"hello"}}},
		{name: "MalformedSumType", input: `{"d":{"message":"hello"}}`, expected: struct{ D discriminator }{}, err: `no name found for type enum variant`},
		{name: "UnregisteredSumType", input: `{"d":{"name":"Variant","value":{"message":"hello"}}}`, expected: struct{ D unregistered }{}, err: `the only supported interface types are enums or any, not encoding_test.unregistered`},
	}

	reflection.AllowAnyPackageForTesting = true
	defer func() { reflection.AllowAnyPackageForTesting = false }()
	reflection.ResetTypeRegistry()
	reflection.Register(reflection.WithSumType[discriminator](variant{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eType := reflect.TypeOf(tt.expected)
			o := reflect.New(eType)
			err := Unmarshal([]byte(tt.input), o.Interface())
			assert.EqualError(t, err, tt.err)
			if err == nil {
				assert.Equal(t, tt.expected, o.Elem().Interface())
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	type inner struct {
		FooBar string
	}
	tests := []struct {
		name  string
		input any
	}{
		{name: "FieldRenaming", input: struct{ FooBar string }{""}},
		{name: "String", input: struct{ String string }{"foo"}},
		{name: "Int", input: struct{ Int int }{42}},
		{name: "Float", input: struct{ Float float64 }{42.42}},
		{name: "Bool", input: struct{ Bool bool }{true}},
		{name: "Slice", input: struct{ Slice []int }{[]int{1, 2, 3}}},
		{name: "SliceOfStrings", input: struct{ Slice []string }{[]string{"hello", "world"}}},
		{name: "Map", input: struct{ Map map[string]int }{map[string]int{"foo": 42}}},
		{name: "Time", input: struct{ Time time.Time }{time.Date(2009, time.November, 29, 21, 33, 0, 0, time.UTC)}},
		{name: "Option", input: struct{ Option ftl.Option[int] }{ftl.Some(42)}},
		{name: "OptionNull", input: struct{ Option ftl.Option[int] }{ftl.None[int]()}},
		{name: "OptionStruct", input: struct{ Option ftl.Option[inner] }{ftl.Some(inner{"foo"})}},
		{name: "Unit", input: ftl.Unit{}},
		{name: "UnitField", input: struct {
			String string
			Unit   ftl.Unit
		}{String: "something", Unit: ftl.Unit{}}},
		{name: "Aliased", input: struct {
			TokenID string `json:"token_id"`
		}{"123"}},
		{name: "SumType", input: struct{ D discriminator }{variant{"hello"}}},
	}

	reflection.AllowAnyPackageForTesting = true
	defer func() { reflection.AllowAnyPackageForTesting = false }()
	reflection.ResetTypeRegistry()
	reflection.Register(reflection.WithSumType[discriminator](variant{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := Marshal(tt.input)
			assert.NoError(t, err)

			eType := reflect.TypeOf(tt.input)
			o := reflect.New(eType)
			err = Unmarshal(marshaled, o.Interface())
			assert.NoError(t, err)

			assert.Equal(t, tt.input, o.Elem().Interface())
		})
	}
}
