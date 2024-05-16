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
						{Name: "Enum", Type: &schema.Ref{Module: "foo", Name: "Color"}},
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
				&schema.Enum{
					Name:   "Color",
					Export: true,
					Type:   &schema.String{},
					Variants: []*schema.EnumVariant{
						{Name: "Red", Value: &schema.StringValue{Value: "Red"}},
						{Name: "Blue", Value: &schema.StringValue{Value: "Blue"}},
						{Name: "Green", Value: &schema.StringValue{Value: "Green"}},
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
  Enum foo.Color
}

data Nested {
  Field String
}

export data BarData {
  Name String
}

export enum Color: String {
  Red = "Red"
  Blue = "Blue"
  Green = "Green"
}

data EchoResponse {
  Message String
}

verb Echo(foo.EchoRequest) foo.EchoResponse`

	schemaString, err := verbSchemaString(sch, verb)
	assert.NoError(t, err)
	assert.Equal(t, expected, schemaString)
}
