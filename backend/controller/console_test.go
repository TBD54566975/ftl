package controller

import (
	"testing"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/alecthomas/assert/v2"
)

func TestVerbSchemaString(t *testing.T) {
	verb := &schema.Verb{
		Name:     "Echo",
		Request:  &schema.Ref{Module: "foo", Name: "EchoRequest"},
		Response: &schema.Ref{Module: "foo", Name: "EchoResponse"},
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			{Name: "foo", Decls: []schema.Decl{
				verb,
				&schema.Data{
					Name: "EchoRequest",
					Fields: []*schema.Field{
						{Name: "Name", Type: &schema.String{}},
						{Name: "Nested", Type: &schema.Ref{Module: "foo", Name: "Nested"}},
						{Name: "External", Type: &schema.Ref{Module: "bar", Name: "BarData"}},
					},
				},
				&schema.Data{
					Name: "EchoResponse",
					Fields: []*schema.Field{
						{Name: "Message", Type: &schema.String{}},
					},
				},
				&schema.Data{
					Name: "Nested",
					Fields: []*schema.Field{
						{Name: "Field", Type: &schema.String{}},
					},
				},
			}},
			{Name: "bar", Decls: []schema.Decl{
				verb,
				&schema.Data{
					Name:   "BarData",
					Export: true,
					Fields: []*schema.Field{
						{Name: "Name", Type: &schema.String{}},
					},
				}},
			}},
	}

	expected := `data EchoRequest {
  Name String
  Nested foo.Nested
  External bar.BarData
}

data Nested {
  Field String
}

export data BarData {
  Name String
}

data EchoResponse {
  Message String
}

verb Echo(foo.EchoRequest) foo.EchoResponse`

	schemaString, err := verbSchemaString(sch, verb)
	assert.NoError(t, err)
	assert.Equal(t, expected, schemaString)
}
