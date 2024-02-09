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
	actual, err := data.Monomorphise(&DataRef{TypeParameters: []Type{&String{}}})
	assert.NoError(t, err)
	expected := &Data{
		Comments: []string{},
		Name:     "Data",
		Fields:   []*Field{{Comments: []string{}, Name: "a", Type: &String{}}},
		Metadata: []Metadata{},
	}
	assert.Equal(t, expected, actual, assert.OmitEmpty())
}
