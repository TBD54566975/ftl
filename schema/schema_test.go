package schema

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"
)

var schema = Schema{
	Modules: []Module{
		{
			Name:     "todo",
			Comments: []string{"A comment"},
			Data: []Data{
				{
					Name: "CreateRequest",
					Fields: []Field{
						{Name: "name", Type: Map{Key: String{Str: true}, Value: String{Str: true}}},
					},
				},
				{
					Name: "CreateResponse",
					Fields: []Field{
						{Name: "name", Type: Array{Element: String{Str: true}}},
					},
				},
				{
					Name: "DestroyRequest",
					Fields: []Field{
						{Name: "name", Comments: []string{"A comment"}, Type: String{Str: true}},
					},
				},
				{
					Name: "DestroyResponse",
					Fields: []Field{
						{Name: "name", Type: String{Str: true}},
					},
				},
			},
			Verbs: []Verb{
				{Name: "create",
					Request:  DataRef{Name: "CreateRequest"},
					Response: DataRef{Name: "CreateResponse"}},
				{Name: "destroy",
					Request:  DataRef{Name: "DestroyRequest"},
					Response: DataRef{Name: "DestroyResponse"},
				},
			},
		},
	},
}

func TestIndent(t *testing.T) {
	assert.Equal(t, "  a\n  b\n  c", indent("a\nb\nc"))
}

func TestSchemaString(t *testing.T) {
	expected := `
// A comment
module todo {
  data CreateRequest {
    name Map<string, string>
  }
  data CreateResponse {
    name Array<string>
  }
  data DestroyRequest {
    // A comment
    name string
  }
  data DestroyResponse {
    name string
  }

  verb create(CreateRequest) CreateResponse
  verb destroy(DestroyRequest) DestroyResponse
}
`
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(schema.String()))
}

func TestVisit(t *testing.T) {
	expected := `
Schema
  Module
    Data
      Field
        Map
          String
          String
    Data
      Field
        Array
          String
    Data
      Field
        String
    Data
      Field
        String
    Verb
      DataRef
      DataRef
    Verb
      DataRef
      DataRef
`
	actual := &strings.Builder{}
	i := 0
	err := Visit(schema, func(n Node, next func() error) error {
		prefix := strings.Repeat(" ", i)
		tn := strings.TrimPrefix(fmt.Sprintf("%T", n), "schema.")
		fmt.Fprintf(actual, "%s%s\n", prefix, tn)
		i += 2
		defer func() { i -= 2 }()
		return next()
	})
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(actual.String()))
}

func TestParserRoundTrip(t *testing.T) {
	actual, err := ParseString("", schema.String())
	assert.NoError(t, err, "%s", schema.String())
	actual = Normalise(actual)
	assert.Equal(t, schema, actual, "%s", schema.String())
}

func TestParser(t *testing.T) {
	actual, err := ParseString("", `
		module todo {
			data CreateListRequest {}
			data CreateListResponse {}

			// Create a new list
			verb createList(CreateListRequest) CreateListResponse
				calls createList
		}
	`)
	assert.NoError(t, err)
	actual = Normalise(actual)
	expected := Schema{
		Modules: []Module{
			{
				Name: "todo",
				Data: []Data{
					{Name: "CreateListRequest"},
					{Name: "CreateListResponse"},
				},
				Verbs: []Verb{
					{Name: "createList",
						Comments: []string{"Create a new list"},
						Request:  DataRef{Name: "CreateListRequest"},
						Response: DataRef{Name: "CreateListResponse"},
						Metadata: []Metadata{
							MetadataCalls{Calls: []VerbRef{{Name: "createList"}}},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expected, actual)
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		errors []string
	}{
		{name: "InvalidRequestRef",
			input: `module test { verb test(InvalidRequest) InvalidResponse}`,
			errors: []string{
				"1:25: reference to unknown Data structure \"InvalidRequest\"",
				"1:41: reference to unknown Data structure \"InvalidResponse\""}},
		{name: "InvalidDataRef",
			input: `module test { data Data { user user.User }}`,
			errors: []string{
				"1:32: reference to unknown Verb \"user.User\""}},
		{name: "InvalidMetadataSyntax",
			input: `module test { data Data {} calls }`,
			errors: []string{
				"1:28: unexpected token \"calls\" (expected \"}\")",
			},
		},
		{name: "InvalidDataMetadata",
			input: `module test { data Data {} calls verb }`,
			errors: []string{
				"1:28: metadata \"calls verb\" is not valid on data structures",
				"1:34: reference to unknown Verb \"verb\"",
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseString("", test.input)
			if test.errors != nil {
				assert.Error(t, err, "expected an error")
				actual := []string{}
				errs := errors.UnwrapAll(err)
				for _, err := range errs {
					if errors.Innermost(err) {
						actual = append(actual, err.Error())
					}
				}
				assert.Equal(t, test.errors, actual)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
