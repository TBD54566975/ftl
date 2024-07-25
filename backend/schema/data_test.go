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
			{Name: "a", Type: &Ref{Name: "T"}},
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
		{typ: &Array{Element: &Ref{Module: "builtin", Name: "Test"}}},
		{typ: &Ref{Module: "builtin", Name: "Test"}},
		{typ: &Ref{Module: "builtin", Name: "Test", TypeParameters: []Type{&String{}}}},
		{typ: &Map{Key: &String{}, Value: &Int{}}},
		{typ: &Map{Key: &Ref{Module: "builtin", Name: "Test"}, Value: &Ref{Module: "builtin", Name: "Test"}}},
		{typ: &Optional{Type: &String{}}},
		{typ: &Optional{Type: &Ref{Module: "builtin", Name: "Test"}}},
		{typ: &Any{}},
	}

	for _, test := range tests {
		actual, err := data.Monomorphise(&Ref{TypeParameters: []Type{test.typ}})
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
