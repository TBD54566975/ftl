package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestMonomorphisation(t *testing.T) {
	data := &Data{
		Name:           "Data",
		TypeParameters: []*TypeParameter{{Name: "T"}},
		Fields: []*Field{
			{Name: "a", Type: &DataRef{Name: "T"}},
		},
	}

	tests := []struct {
		typ Type
	}{
		{typ: &String{}},
		{typ: &Int{}},
		{typ: &Float{}},
		{typ: &Bool{}},
		{typ: &Array{Element: &String{}}},
		{typ: &Array{Element: &DataRef{Module: "builtin", Name: "Test"}}},
		{typ: &DataRef{Module: "builtin", Name: "Test"}},
		{typ: &DataRef{Module: "builtin", Name: "Test", TypeParameters: []Type{&String{}}}},
		{typ: &Map{Key: &String{}, Value: &Int{}}},
		{typ: &Map{Key: &DataRef{Module: "builtin", Name: "Test"}, Value: &DataRef{Module: "builtin", Name: "Test"}}},
		{typ: &Optional{Type: &String{}}},
		{typ: &Optional{Type: &DataRef{Module: "builtin", Name: "Test"}}},
		{typ: &Any{}},
	}

	for _, test := range tests {
		actual, err := data.Monomorphise(&DataRef{TypeParameters: []Type{test.typ}})
		assert.NoError(t, err)
		expected := &Data{
			Comments: []string{},
			Name:     "Data",
			Fields:   []*Field{{Comments: []string{}, Name: "a", Type: test.typ}},
			Metadata: []Metadata{},
		}
		assert.Equal(t, expected, actual, assert.OmitEmpty())
	}
}
