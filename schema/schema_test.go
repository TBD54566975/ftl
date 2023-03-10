package schema

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

var schema = Module{
	Name: "todo",
	Data: []Data{
		{
			Name: "CreateRequest",
			Fields: []Field{
				{Name: "name", Type: Map{String{}, String{}}},
			},
		},
		{
			Name: "CreateResponse",
			Fields: []Field{
				{Name: "name", Type: Array{Element: String{}}},
			},
		},
		{
			Name: "DestroyRequest",
			Fields: []Field{
				{Name: "name", Type: String{}},
			},
		},
		{
			Name: "DestroyResponse",
			Fields: []Field{
				{Name: "name", Type: String{}},
			},
		},
	},
	Verbs: []Verb{
		{Name: "create",
			Request:  DataRef{"CreateRequest"},
			Response: DataRef{"CreateResponse"}},
		{Name: "destroy",
			Request:  DataRef{"DestroyRequest"},
			Response: DataRef{"DestroyResponse"},
			Calls:    []VerbRef{{Module: "notify", Verb: "user"}},
		},
	},
}

func TestIndent(t *testing.T) {
	assert.Equal(t, "  a\n  b\n  c", indent("a\nb\nc", "  "))
}

func TestSchemaString(t *testing.T) {
	fmt.Println(schema.String())
}

func TestVisit(t *testing.T) {
	i := 0
	err := Visit(schema, func(n Node, next func() error) error {
		prefix := strings.Repeat(" ", i)
		tn := strings.TrimPrefix(fmt.Sprintf("%T", n), "schema.")
		fmt.Printf("%s%s\n", prefix, tn)
		i += 2
		defer func() { i -= 2 }()
		return next()
	})
	assert.NoError(t, err)
}

func TestMarshalUnmarshalJSONRoundTrip(t *testing.T) {
	b, err := json.MarshalIndent(schema, "", "  ")
	assert.NoError(t, err)
	actual := Module{}
	err = json.Unmarshal(b, &actual)
	assert.NoError(t, err)
	assert.Equal(t, schema, actual)
	t.Log(string(b))
}
