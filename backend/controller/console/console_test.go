package console

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
)

func TestVerbSchemaString(t *testing.T) {
	verb := &schema.Verb{
		Name:     "Echo",
		Request:  &schema.Ref{Module: "foo", Name: "EchoRequest"},
		Response: &schema.Ref{Module: "foo", Name: "EchoResponse"},
	}
	ingressVerb := &schema.Verb{
		Name:     "Ingress",
		Request:  &schema.Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []schema.Type{&schema.String{}, &schema.Unit{}, &schema.Unit{}}},
		Response: &schema.Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []schema.Type{&schema.String{}, &schema.String{}}},
		Metadata: []schema.Metadata{
			&schema.MetadataIngress{Type: "http", Method: "GET", Path: []schema.IngressPathComponent{&schema.IngressPathLiteral{Text: "test"}}},
		},
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "foo", Decls: []schema.Decl{
				verb,
				ingressVerb,
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
				ingressVerb,
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

func TestVerbSchemaStringIngress(t *testing.T) {
	verb := &schema.Verb{
		Name:     "Ingress",
		Request:  &schema.Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []schema.Type{&schema.Ref{Module: "foo", Name: "FooRequest"}, &schema.Unit{}, &schema.Unit{}}},
		Response: &schema.Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []schema.Type{&schema.Ref{Module: "foo", Name: "FooResponse"}, &schema.String{}}},
		Metadata: []schema.Metadata{
			&schema.MetadataIngress{Type: "http", Method: "GET", Path: []schema.IngressPathComponent{&schema.IngressPathLiteral{Text: "foo"}}},
		},
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "foo", Decls: []schema.Decl{
				verb,
				&schema.Data{
					Name: "FooRequest",
					Fields: []*schema.Field{
						{Name: "Name", Type: &schema.String{}},
					},
				},
				&schema.Data{
					Name: "FooResponse",
					Fields: []*schema.Field{
						{Name: "Message", Type: &schema.String{}},
					},
				},
			}},
		},
	}

	expected := `// HTTP request structure used for HTTP ingress verbs.
export data HttpRequest<Body, Path, Query> {
  method String
  path String
  pathParameters Path
  query Query
  headers {String: [String]}
  body Body
}

data FooRequest {
  Name String
}

// HTTP response structure used for HTTP ingress verbs.
export data HttpResponse<Body, Error> {
  status Int
  headers {String: [String]}
  // Either "body" or "error" must be present, not both.
  body Body?
  error Error?
}

data FooResponse {
  Message String
}

verb Ingress(builtin.HttpRequest<foo.FooRequest, Unit, Unit>) builtin.HttpResponse<foo.FooResponse, String>  
  +ingress http GET /foo`

	schemaString, err := verbSchemaString(sch, verb)
	assert.NoError(t, err)
	assert.Equal(t, expected, schemaString)
}
