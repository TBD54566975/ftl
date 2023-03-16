package schema

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
)

type TestRequest struct{ ID string }
type TestResponse struct{ ID string }

func testVerb(ctx context.Context, req TestRequest) (TestResponse, error) { panic("not implemented") }

func TestReflectVerb(t *testing.T) {
	module := Module{Name: "github.com/TBD54566975/ftl"}
	err := ReflectVerbIntoModule(&module, testVerb)
	assert.NoError(t, err)
	expected := Module{
		Name: "github.com/TBD54566975/ftl",
		Verbs: []Verb{
			{
				Name:     "schema/testVerb",
				Request:  DataRef{Name: "schema/TestRequest"},
				Response: DataRef{Name: "schema/TestResponse"},
			},
		},
		Data: []Data{
			{
				Name: "schema/TestRequest",
				Fields: []Field{
					{Name: "ID", Type: String{Str: true}},
				},
			},
			{
				Name: "schema/TestResponse",
				Fields: []Field{
					{Name: "ID", Type: String{Str: true}},
				},
			},
		},
	}
	assert.Equal(t, expected, module)
	t.Log(module.String())
}
