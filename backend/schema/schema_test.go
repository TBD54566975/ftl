package schema

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/internal/errors"
)

func TestIndent(t *testing.T) {
	assert.Equal(t, "  a\n  b\n  c", indent("a\nb\nc"))
}

func TestSchemaString(t *testing.T) {
	expected := Builtins().String() + `
// A comment
module todo {
  data CreateRequest {
    name {String: String}? alias rqn
  }

  data CreateResponse {
    name [String] alias rsn
  }

  data DestroyRequest {
    // A comment
    name String
  }

  data DestroyResponse {
    name String
    when Time
  }

  verb create(todo.CreateRequest) todo.CreateResponse
      calls todo.destroy


  verb destroy(todo.DestroyRequest) todo.DestroyResponse
      ingress ftl GET /todo/destroy/{id}
}
`
	assert.Equal(t, normaliseString(expected), normaliseString(testSchema.String()))
}

func normaliseString(s string) string {
	return strings.TrimSpace(strings.Join(slices.Map(strings.Split(s, "\n"), strings.TrimSpace), "\n"))
}

func TestImports(t *testing.T) {
	input := `
	module test {
		data Data {
			ref other.Data
			ref another.Data
		}
	}
	`
	schema, err := ParseModuleString("", input)
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "other"}, schema.Imports())
}

func TestVisit(t *testing.T) {
	expected := `
Module
  Data
    Field
      Optional
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
    Field
      Time
  Verb
    DataRef
    DataRef
    MetadataCalls
      VerbRef
  Verb
    DataRef
    DataRef
    MetadataIngress
      IngressPathLiteral
      IngressPathLiteral
      IngressPathParameter
`
	actual := &strings.Builder{}
	i := 0
	// Modules[0] is always the builtins, which we skip.
	err := Visit(testSchema.Modules[1], func(n Node, next func() error) error {
		prefix := strings.Repeat(" ", i)
		fmt.Fprintf(actual, "%s%s\n", prefix, TypeName(n))
		i += 2
		defer func() { i -= 2 }()
		return next()
	})
	assert.NoError(t, err)
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()), "%s", actual.String())
}

func TestParserRoundTrip(t *testing.T) {
	actual, err := ParseString("", testSchema.String())
	assert.NoError(t, err, "%s", testSchema.String())
	actual, err = Validate(actual)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, Normalise(testSchema), Normalise(actual), "%s", testSchema.String())
}

func TestParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		errors   []string
		expected *Schema
	}{
		{name: "Example",
			input: `
				module todo {
					data CreateListRequest {}
					data CreateListResponse {}

					// Create a new list
					verb createList(todo.CreateListRequest) CreateListResponse
						calls createList
				}
			`,
			expected: &Schema{
				Modules: []*Module{
					{
						Name: "todo",
						Decls: []Decl{
							&Data{Name: "CreateListRequest"},
							&Data{Name: "CreateListResponse"},
							&Verb{Name: "createList",
								Comments: []string{"Create a new list"},
								Request:  &DataRef{Module: "todo", Name: "CreateListRequest"},
								Response: &DataRef{Module: "todo", Name: "CreateListResponse"},
								Metadata: []Metadata{
									&MetadataCalls{Calls: []*VerbRef{{Module: "todo", Name: "createList"}}},
								},
							},
						},
					},
				},
			}},
		{name: "InvalidRequestRef",
			input: `module test { verb test(InvalidRequest) InvalidResponse}`,
			errors: []string{
				"1:25: reference to unknown data structure \"InvalidRequest\"",
				"1:41: reference to unknown data structure \"InvalidResponse\""}},
		{name: "InvalidDataRef",
			input: `module test { data Data { user user.User }}`,
			errors: []string{
				"1:32: reference to unknown data structure \"user.User\""}},
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
				"1:34: reference to unknown verb \"verb\"",
			}},
		{name: "KeywordAsName",
			input:  `module int { data String { name String } verb verb(String) String }`,
			errors: []string{"1:14: data structure name \"String\" is a reserved word"}},
		{name: "BuiltinRef",
			input: `module test { verb myIngress(HttpRequest<string>) HttpResponse<string> }`,
			expected: &Schema{
				Modules: []*Module{{
					Name: "test",
					Decls: []Decl{
						&Verb{
							Name: "myIngress",
							Request: &DataRef{Module: "builtin", Name: "HttpRequest", TypeParameters: []Type{
								&DataRef{
									Pos: Position{
										Offset: 41,
										Line:   1,
										Column: 42,
									},
									Name:           "string",
									TypeParameters: []Type{},
								},
							}},
							Response: &DataRef{Module: "builtin", Name: "HttpResponse", TypeParameters: []Type{
								&DataRef{
									Pos: Position{
										Offset: 63,
										Line:   1,
										Column: 64,
									},
									Name:           "string",
									TypeParameters: []Type{},
								},
							}},
						},
					},
				}},
			},
		},
		{name: "TimeEcho",
			input: `
				module echo {
					data EchoRequest {
						name String?
					}

					data EchoResponse {
						message String
					}

					verb echo(echo.EchoRequest) echo.EchoResponse
						ingress ftl GET /echo
						calls time.time

				}

				module time {
					data TimeRequest {
					}

					data TimeResponse {
						time Time
					}

					verb time(time.TimeRequest) time.TimeResponse
						ingress ftl GET /time
				}
				`,
			expected: &Schema{
				Modules: []*Module{{
					Name: "echo",
					Decls: []Decl{
						&Data{Name: "EchoRequest", Fields: []*Field{{Name: "name", Type: &Optional{Type: &String{}}}}},
						&Data{Name: "EchoResponse", Fields: []*Field{{Name: "message", Type: &String{}}}},
						&Verb{
							Name:     "echo",
							Request:  &DataRef{Module: "echo", Name: "EchoRequest"},
							Response: &DataRef{Module: "echo", Name: "EchoResponse"},
							Metadata: []Metadata{
								&MetadataIngress{Type: "ftl", Method: "GET", Path: []IngressPathComponent{&IngressPathLiteral{Text: "echo"}}},
								&MetadataCalls{Calls: []*VerbRef{{Module: "time", Name: "time"}}},
							},
						},
					},
				}, {
					Name: "time",
					Decls: []Decl{
						&Data{Name: "TimeRequest"},
						&Data{Name: "TimeResponse", Fields: []*Field{{Name: "time", Type: &Time{}}}},
						&Verb{
							Name:     "time",
							Request:  &DataRef{Module: "time", Name: "TimeRequest"},
							Response: &DataRef{Module: "time", Name: "TimeResponse"},
							Metadata: []Metadata{
								&MetadataIngress{Type: "ftl", Method: "GET", Path: []IngressPathComponent{&IngressPathLiteral{Text: "time"}}},
							},
						},
					},
				}},
			},
		},
		{name: "TypeParameters",
			input: `
				module test {
					data Data<T> {
						value T
					}

					verb test(Data<String>) Data<String>
				}
				`,
			expected: &Schema{
				Modules: []*Module{{
					Name: "test",
					Decls: []Decl{
						&Data{
							Comments:       []string{},
							Name:           "Data",
							TypeParameters: []*TypeParameter{{Name: "T"}},
							Fields: []*Field{
								{Name: "value", Type: &DataRef{Name: "T", TypeParameters: []Type{}}},
							},
						},
						&Verb{
							Comments: []string{},
							Name:     "test",
							Request: &DataRef{
								Module: "test",
								Name:   "Data",
								TypeParameters: []Type{
									&String{
										Pos: Position{
											Offset: 81,
											Line:   7,
											Column: 21,
										},
										Str: true,
									},
								},
							},
							Response: &DataRef{
								Module: "test",
								Name:   "Data",
								TypeParameters: []Type{
									&String{
										Pos: Position{
											Offset: 95,
											Line:   7,
											Column: 35,
										},
										Str: true,
									},
								},
							},
						},
					},
				}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseString("", test.input)
			if test.errors != nil {
				assert.Error(t, err, "expected errors")
				actual := []string{}
				errs := errors.UnwrapAll(err)
				for _, err := range errs {
					if errors.Innermost(err) {
						actual = append(actual, err.Error())
					}
				}
				assert.Equal(t, test.errors, actual, test.input)
			} else {
				assert.NoError(t, err)
				actual = Normalise(actual)
				assert.NotZero(t, test.expected, "test.expected is nil")
				assert.NotZero(t, test.expected.Modules, "test.expected.Modules is nil")
				test.expected.Modules = append([]*Module{Builtins()}, test.expected.Modules...)
				assert.Equal(t, Normalise(test.expected), Normalise(actual), test.input, assert.OmitEmpty())
			}
		})
	}
}

func TestParseModule(t *testing.T) {
	input := `
// A comment
module todo {
  data CreateRequest {
    name {String: String}? alias rqn
  }
  data CreateResponse {
    name [String] alias rsn
  }
  data DestroyRequest {
    // A comment
    name String
  }
  data DestroyResponse {
    name String
	when Time
  }
  verb create(todo.CreateRequest) todo.CreateResponse
  	calls todo.destroy
  verb destroy(todo.DestroyRequest) todo.DestroyResponse
  	ingress ftl GET /todo/destroy/{id}
}
`
	actual, err := ParseModuleString("", input)
	assert.NoError(t, err)
	actual = Normalise(actual)
	fmt.Printf("Modules %v\n", Normalise(testSchema.Modules[1]))
	fmt.Printf("Modules %v\n", Normalise(actual))
	assert.Equal(t, Normalise(testSchema.Modules[1]), actual)
}

var testSchema = MustValidate(&Schema{
	Modules: []*Module{
		{
			Name:     "todo",
			Comments: []string{"A comment"},
			Decls: []Decl{
				&Data{
					Name: "CreateRequest",
					Fields: []*Field{
						{Name: "name", Type: &Optional{Type: &Map{Key: &String{}, Value: &String{}}}, Alias: "rqn"},
					},
				},
				&Data{
					Name: "CreateResponse",
					Fields: []*Field{
						{Name: "name", Type: &Array{Element: &String{}}, Alias: "rsn"},
					},
				},
				&Data{
					Name: "DestroyRequest",
					Fields: []*Field{
						{Name: "name", Comments: []string{"A comment"}, Type: &String{}},
					},
				},
				&Data{
					Name: "DestroyResponse",
					Fields: []*Field{
						{Name: "name", Type: &String{}},
						{Name: "when", Type: &Time{}},
					},
				},
				&Verb{Name: "create",
					Request:  &DataRef{Module: "todo", Name: "CreateRequest"},
					Response: &DataRef{Module: "todo", Name: "CreateResponse"},
					Metadata: []Metadata{&MetadataCalls{Calls: []*VerbRef{{Module: "todo", Name: "destroy"}}}}},
				&Verb{Name: "destroy",
					Request:  &DataRef{Module: "todo", Name: "DestroyRequest"},
					Response: &DataRef{Module: "todo", Name: "DestroyResponse"},
					Metadata: []Metadata{
						&MetadataIngress{
							Type:   "ftl",
							Method: "GET",
							Path: []IngressPathComponent{
								&IngressPathLiteral{Text: "todo"},
								&IngressPathLiteral{Text: "destroy"},
								&IngressPathParameter{Name: "id"},
							},
						},
					},
				},
			},
		},
	},
})
