package schema

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/slices"
)

func TestIndent(t *testing.T) {
	assert.Equal(t, "  a\n  b\n  c", indent("a\nb\nc"))
}

func TestSchemaString(t *testing.T) {
	expected := Builtins().String() + `
// A comment
module todo {
  // A config value
  config configValue String
  // Shhh
  secret secretValue String

  // A database
  database postgres testdb

  export data CreateRequest {
    name {String: String}? +alias json "rqn"
  }

  export data CreateResponse {
    name [String] +alias json "rsn"
  }

  data DestroyRequest {
    // A comment
    name String
  }

  data DestroyResponse {
    name String
    when Time
  }

  export data PersonalInfo {
    name Encrypted<String>
    age Encrypted<Int>
    photo Encrypted<Bytes>
  }

  export verb create(todo.CreateRequest) todo.CreateResponse
    +calls todo.destroy
	+database calls todo.testdb

  export verb destroy(builtin.HttpRequest<Unit, todo.DestroyRequest, Unit>) builtin.HttpResponse<todo.DestroyResponse, String>
      +ingress http GET /todo/destroy/{name}

  verb mondays(Unit) Unit
      +cron Mon

  verb scheduled(Unit) Unit
      +cron */10 * * 1-10,11-31 * * *

  verb twiceADay(Unit) Unit
      +cron 12h
}

module foo {
  // A comment
  enum Color: String {
	Red = "Red"
	Blue = "Blue"
	Green = "Green"
  }

  export enum ColorInt: Int {
	Red = 0
	Blue = 1
	Green = 2
  }

  enum StringTypeEnum {
	A String
	B String
  }

  enum TypeEnum {
	A String
	B [String]
	C Int
  }

  verb callTodoCreate(todo.CreateRequest) todo.CreateResponse
      +calls todo.create
}

module payments {
  fsm payment {
    start payments.created
    start payments.paid
    transition payments.created to payments.paid
    transition payments.created to payments.failed
    transition payments.paid to payments.completed
  }

  data OnlinePaymentCompleted {
  }

  data OnlinePaymentCreated {
  }

  data OnlinePaymentFailed {
  }

  data OnlinePaymentPaid {
  }

  verb completed(payments.OnlinePaymentCompleted) Unit

  verb created(payments.OnlinePaymentCreated) Unit

  verb failed(payments.OnlinePaymentFailed) Unit

  verb paid(payments.OnlinePaymentPaid) Unit
}

module typealias {
  typealias NonFtlType Any
      +typemap go "github.com/foo/bar.Type"
      +typemap kotlin "com.foo.bar.Type"
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
		data Generic<T> {
			value T
		}
		data Data {
			ref other.Data
			ref another.Data
			ref test.Generic<new.Data>
		}
		verb myVerb(test.Data) test.Data
			+calls verbose.verb
	}
	`
	schema, err := ParseModuleString("", input)
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "new", "other", "verbose"}, schema.Imports())
}

func TestVisit(t *testing.T) {
	expected := `
Module
  Config
    String
  Secret
    String
  Database
  Data
    Field
      Optional
        Map
          String
          String
      MetadataAlias
  Data
    Field
      Array
        String
      MetadataAlias
  Data
    Field
      String
  Data
    Field
      String
    Field
      Time
  Verb
    Ref
    Ref
    MetadataCalls
      Ref
    MetadataDatabases
      Ref
  Verb
    Ref
      Unit
      Ref
      Unit
    Ref
      Ref
      String
    MetadataIngress
      IngressPathLiteral
      IngressPathLiteral
      IngressPathParameter
  Verb
    Unit
    Unit
    MetadataCronJob
  Verb
    Unit
    Unit
    MetadataCronJob
  Verb
    Unit
    Unit
    MetadataCronJob
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
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(actual.String()))
}

func TestParserRoundTrip(t *testing.T) {
	actual, err := ParseString("", testSchema.String())
	assert.NoError(t, err, "%s", testSchema.String())
	actual, err = ValidateSchema(actual)
	assert.NoError(t, err)
	assert.Equal(t, Normalise(testSchema), Normalise(actual), assert.Exclude[Position]())
}

//nolint:maintidx
func TestParsing(t *testing.T) {
	zero := 0
	ten := 10
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
					verb createList(todo.CreateListRequest) todo.CreateListResponse
						+calls todo.createList
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
								Request:  &Ref{Module: "todo", Name: "CreateListRequest"},
								Response: &Ref{Module: "todo", Name: "CreateListResponse"},
								Metadata: []Metadata{
									&MetadataCalls{Calls: []*Ref{{Module: "todo", Name: "createList"}}},
								},
							},
						},
					},
				},
			}},
		{name: "InvalidRequestRef",
			input: `module test { verb test(InvalidRequest) InvalidResponse}`,
			errors: []string{
				"1:25-25: unknown reference \"InvalidRequest\", is the type annotated and exported?",
				"1:41-41: unknown reference \"InvalidResponse\", is the type annotated and exported?"}},
		{name: "InvalidRef",
			input: `module test { data Data { user user.User }}`,
			errors: []string{
				"1:32-32: unknown reference \"user.User\", is the type annotated and exported?"}},
		{name: "InvalidMetadataSyntax",
			input: `module test { data Data {} calls }`,
			errors: []string{
				"1:28: unexpected token \"calls\" (expected \"}\")",
			},
		},
		{name: "InvalidDataMetadata",
			input: `module test { data Data {} +calls verb }`,
			errors: []string{
				"1:28-28: metadata \"+calls verb\" is not valid on data structures",
				"1:35-35: unknown reference \"verb\", is the type annotated and exported?",
			}},
		{name: "KeywordAsName",
			input:  `module int { data String { name String } verb verb(String) String }`,
			errors: []string{"1:14-14: data name \"String\" is a reserved word"}},
		{name: "BuiltinRef",
			input: `module test { verb myIngress(HttpRequest<String, Unit, Unit>) HttpResponse<String, String> }`,
			expected: &Schema{
				Modules: []*Module{{
					Name: "test",
					Decls: []Decl{
						&Verb{
							Name:     "myIngress",
							Request:  &Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []Type{&String{}, &Unit{}, &Unit{}}},
							Response: &Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []Type{&String{}, &String{}}},
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

					export verb echo(builtin.HttpRequest<Unit, Unit, echo.EchoRequest>) builtin.HttpResponse<echo.EchoResponse, String>
						+ingress http GET /echo
						+calls time.time

				}

				module time {
					data TimeRequest {
					}

					data TimeResponse {
						time Time
					}

					export verb time(builtin.HttpRequest<Unit, Unit, Unit>) builtin.HttpResponse<time.TimeResponse, String>
						+ingress http GET /time
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
							Export:   true,
							Request:  &Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []Type{&Unit{}, &Unit{}, &Ref{Module: "echo", Name: "EchoRequest"}}},
							Response: &Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []Type{&Ref{Module: "echo", Name: "EchoResponse"}, &String{}}},
							Metadata: []Metadata{
								&MetadataIngress{Type: "http", Method: "GET", Path: []IngressPathComponent{&IngressPathLiteral{Text: "echo"}}},
								&MetadataCalls{Calls: []*Ref{{Module: "time", Name: "time"}}},
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
							Export:   true,
							Request:  &Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []Type{&Unit{}, &Unit{}, &Unit{}}},
							Response: &Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []Type{&Ref{Module: "time", Name: "TimeResponse"}, &String{}}},
							Metadata: []Metadata{
								&MetadataIngress{Type: "http", Method: "GET", Path: []IngressPathComponent{&IngressPathLiteral{Text: "time"}}},
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

					verb test(test.Data<String>) test.Data<String>
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
								{Name: "value", Type: &Ref{Name: "T", TypeParameters: []Type{}}},
							},
						},
						&Verb{
							Comments: []string{},
							Name:     "test",
							Request: &Ref{
								Module:         "test",
								Name:           "Data",
								TypeParameters: []Type{&String{}},
							},
							Response: &Ref{
								Module:         "test",
								Name:           "Data",
								TypeParameters: []Type{&String{}},
							},
						},
					},
				}},
			},
		},
		{name: "RetryFSM",
			input: `
				module test {
					verb A(Empty) Unit
						+retry 10 1m5s 90s
					verb B(Empty) Unit
						+retry 1h1m5s
					verb C(Empty) Unit
						+retry 0h0m5s 1h0m0s
					verb D(Empty) Unit
						+retry 0
					verb E(Empty) Unit
						+retry 0 1s catch test.catch
					verb F(Empty) Unit
						+retry 0 catch test.catch
					verb catch(builtin.CatchRequest<Any>) Unit

					fsm FSM
						+ retry 10 1s 10s
					{
						start test.A
						transition test.A to test.B
						transition test.B to test.C
						transition test.C to test.D
						transition test.D to test.E
						transition test.E to test.F
					}
				}
				`,
			expected: &Schema{
				Modules: []*Module{{
					Name: "test",
					Decls: []Decl{
						&FSM{
							Name: "FSM",
							Metadata: []Metadata{
								&MetadataRetry{
									Count:      &ten,
									MinBackoff: "1s",
									MaxBackoff: "10s",
								},
							},
							Start: []*Ref{
								{
									Module: "test",
									Name:   "A",
								},
							},
							Transitions: []*FSMTransition{
								{
									From: &Ref{
										Module: "test",
										Name:   "A",
									},
									To: &Ref{
										Module: "test",
										Name:   "B",
									},
								},
								{
									From: &Ref{
										Module: "test",
										Name:   "B",
									},
									To: &Ref{
										Module: "test",
										Name:   "C",
									},
								},
								{
									From: &Ref{
										Module: "test",
										Name:   "C",
									},
									To: &Ref{
										Module: "test",
										Name:   "D",
									},
								},
								{
									From: &Ref{
										Module: "test",
										Name:   "D",
									},
									To: &Ref{
										Module: "test",
										Name:   "E",
									},
								},
								{
									From: &Ref{
										Module: "test",
										Name:   "E",
									},
									To: &Ref{
										Module: "test",
										Name:   "F",
									},
								},
							},
						},
						&Verb{
							Comments: []string{},
							Name:     "A",
							Request: &Ref{
								Module: "builtin",
								Name:   "Empty",
							},
							Response: &Unit{
								Unit: true,
							},
							Metadata: []Metadata{
								&MetadataRetry{
									Count:      &ten,
									MinBackoff: "1m5s",
									MaxBackoff: "90s",
								},
							},
						},
						&Verb{
							Comments: []string{},
							Name:     "B",
							Request: &Ref{
								Module: "builtin",
								Name:   "Empty",
							},
							Response: &Unit{
								Unit: true,
							},
							Metadata: []Metadata{
								&MetadataRetry{
									Count:      nil,
									MinBackoff: "1h1m5s",
									MaxBackoff: "",
								},
							},
						},
						&Verb{
							Comments: []string{},
							Name:     "C",
							Request: &Ref{
								Module: "builtin",
								Name:   "Empty",
							},
							Response: &Unit{
								Unit: true,
							},
							Metadata: []Metadata{
								&MetadataRetry{
									Count:      nil,
									MinBackoff: "0h0m5s",
									MaxBackoff: "1h0m0s",
								},
							},
						},
						&Verb{
							Name: "D",
							Request: &Ref{
								Module: "builtin",
								Name:   "Empty",
							},
							Response: &Unit{
								Unit: true,
							},
							Metadata: []Metadata{
								&MetadataRetry{
									Count: &zero,
								},
							},
						},
						&Verb{
							Name: "E",
							Request: &Ref{
								Module: "builtin",
								Name:   "Empty",
							},
							Response: &Unit{
								Unit: true,
							},
							Metadata: []Metadata{
								&MetadataRetry{
									Count:      &zero,
									MinBackoff: "1s",
									Catch: &Ref{
										Module: "test",
										Name:   "catch",
									},
								},
							},
						},
						&Verb{
							Name: "F",
							Request: &Ref{
								Module: "builtin",
								Name:   "Empty",
							},
							Response: &Unit{
								Unit: true,
							},
							Metadata: []Metadata{
								&MetadataRetry{
									Count: &zero,
									Catch: &Ref{
										Module: "test",
										Name:   "catch",
									},
								},
							},
						},
						&Verb{
							Name: "catch",
							Request: &Ref{
								Module: "builtin",
								Name:   "CatchRequest",
								TypeParameters: []Type{
									&Any{},
								},
							},
							Response: &Unit{
								Unit: true,
							},
						},
					},
				}},
			},
		},
		{name: "PubSub",
			input: `
				module test {
					export topic topicA test.eventA

					topic topicB test.eventB

					subscription subA1 test.topicA

					subscription subA2 test.topicA

					subscription subB test.topicB

					export data eventA {
					}

					data eventB {
					}

					verb consumesA(test.eventA) Unit
						+subscribe subA1
						+retry 1m5s 1h catch catchesAny

					verb consumesB1(test.eventB) Unit
						+subscribe subB
						+retry 1m5s 1h catch catchesB

					verb consumesBothASubs(test.eventA) Unit
						+subscribe subA1
						+subscribe subA2
						+retry 1m5s 1h catch test.catchesA

					verb catchesA(builtin.CatchRequest<test.eventA>) Unit

					verb catchesB(builtin.CatchRequest<test.eventB>) Unit

					verb catchesAny(builtin.CatchRequest<Any>) Unit
				}
			`,
			expected: &Schema{
				Modules: []*Module{{
					Name: "test",
					Decls: []Decl{
						&Topic{
							Export: true,
							Name:   "topicA",
							Event: &Ref{
								Module: "test",
								Name:   "eventA",
							},
						},
						&Topic{
							Name: "topicB",
							Event: &Ref{
								Module: "test",
								Name:   "eventB",
							},
						},
						&Subscription{
							Name: "subA1",
							Topic: &Ref{
								Module: "test",
								Name:   "topicA",
							},
						},
						&Subscription{
							Name: "subA2",
							Topic: &Ref{
								Module: "test",
								Name:   "topicA",
							},
						},
						&Subscription{
							Name: "subB",
							Topic: &Ref{
								Module: "test",
								Name:   "topicB",
							},
						},
						&Data{
							Export: true,
							Name:   "eventA",
						},
						&Data{
							Name: "eventB",
						},
						&Verb{
							Name: "catchesA",
							Request: &Ref{
								Module: "builtin",
								Name:   "CatchRequest",
								TypeParameters: []Type{
									&Ref{
										Module: "test",
										Name:   "eventA",
									},
								},
							},
							Response: &Unit{
								Unit: true,
							},
						},
						&Verb{
							Name: "catchesAny",
							Request: &Ref{
								Module: "builtin",
								Name:   "CatchRequest",
								TypeParameters: []Type{
									&Any{},
								},
							},
							Response: &Unit{
								Unit: true,
							},
						},
						&Verb{
							Name: "catchesB",
							Request: &Ref{
								Module: "builtin",
								Name:   "CatchRequest",
								TypeParameters: []Type{
									&Ref{
										Module: "test",
										Name:   "eventB",
									},
								},
							},
							Response: &Unit{
								Unit: true,
							},
						},
						&Verb{
							Name: "consumesA",
							Request: &Ref{
								Module: "test",
								Name:   "eventA",
							},
							Response: &Unit{
								Unit: true,
							},
							Metadata: []Metadata{
								&MetadataSubscriber{
									Name: "subA1",
								},
								&MetadataRetry{
									MinBackoff: "1m5s",
									MaxBackoff: "1h",
									Catch: &Ref{
										Module: "test",
										Name:   "catchesAny",
									},
								},
							},
						},
						&Verb{
							Name: "consumesB1",
							Request: &Ref{
								Module: "test",
								Name:   "eventB",
							},
							Response: &Unit{
								Unit: true,
							},
							Metadata: []Metadata{
								&MetadataSubscriber{
									Name: "subB",
								},
								&MetadataRetry{
									MinBackoff: "1m5s",
									MaxBackoff: "1h",
									Catch: &Ref{
										Module: "test",
										Name:   "catchesB",
									},
								},
							},
						},
						&Verb{
							Name: "consumesBothASubs",
							Request: &Ref{
								Module: "test",
								Name:   "eventA",
							},
							Response: &Unit{
								Unit: true,
							},
							Metadata: []Metadata{
								&MetadataSubscriber{
									Name: "subA1",
								},
								&MetadataSubscriber{
									Name: "subA2",
								},
								&MetadataRetry{
									MinBackoff: "1m5s",
									MaxBackoff: "1h",
									Catch: &Ref{
										Module: "test",
										Name:   "catchesA",
									},
								},
							},
						},
					}},
				},
			},
		},
		{name: "PubSubErrors",
			input: `
				module test {
					export topic topicA test.eventA

					subscription subA test.topicB

					export data eventA {
					}

					verb consumesB(test.eventB) Unit
						+subscribe subB
				}
			`,
			errors: []string{
				`10:21-21: unknown reference "test.eventB", is the type annotated and exported?`,
				`11:7-7: verb consumesB: could not find subscription "subB"`,
				`5:24-24: unknown reference "test.topicB", is the type annotated and exported?`,
			},
		},
		{name: "Cron",
			input: `
				module test {
					verb A(Unit) Unit
						+cron Wed
					verb B(Unit) Unit
						+cron */10 * * * * * *
					verb C(Unit) Unit
						+cron 12h
				}
			`,
			expected: &Schema{
				Modules: []*Module{{
					Name: "test",
					Decls: []Decl{
						&Verb{
							Name:     "A",
							Request:  &Unit{Unit: true},
							Response: &Unit{Unit: true},
							Metadata: []Metadata{
								&MetadataCronJob{
									Cron: "Wed",
								},
							},
						},
						&Verb{
							Name:     "B",
							Request:  &Unit{Unit: true},
							Response: &Unit{Unit: true},
							Metadata: []Metadata{
								&MetadataCronJob{
									Cron: "*/10 * * * * * *",
								},
							},
						},
						&Verb{
							Name:     "C",
							Request:  &Unit{Unit: true},
							Response: &Unit{Unit: true},
							Metadata: []Metadata{
								&MetadataCronJob{
									Cron: "12h",
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
				assert.Equal(t, Normalise(test.expected), Normalise(actual), assert.OmitEmpty(), assert.Exclude[Position]())
			}
		})
	}
}

func TestParseModule(t *testing.T) {
	input := `
// A comment
module todo {
  // A config value
  config configValue String
  // Shhh
  secret secretValue String
  // A database
  database postgres testdb

  export data CreateRequest {
    name {String: String}? +alias json "rqn"
  }
  export data CreateResponse {
    name [String] +alias json "rsn"
  }
  data DestroyRequest {
    // A comment
    name String
  }
  data DestroyResponse {
    name String
	when Time
  }
  export verb create(todo.CreateRequest) todo.CreateResponse
  	+calls todo.destroy +database calls todo.testdb
  export verb destroy(builtin.HttpRequest<Unit, todo.DestroyRequest, Unit>) builtin.HttpResponse<todo.DestroyResponse, String>
  	+ingress http GET /todo/destroy/{name}
  verb scheduled(Unit) Unit
    +cron */10 * * 1-10,11-31 * * *
  verb twiceADay(Unit) Unit
    +cron 12h
  verb mondays(Unit) Unit
    +cron Mon
}
`
	actual, err := ParseModuleString("", input)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, Normalise(testSchema.Modules[1]), actual, assert.Exclude[Position]())
}

func TestParseEnum(t *testing.T) {
	input := `
	module foo {
	 // A comment
	 enum Color: String {
		Red = "Red"
		Blue = "Blue"
		Green = "Green"
	 }

	 export enum ColorInt: Int {
		Red = 0
		Blue = 1
		Green = 2
	 }

	 enum TypeEnum {
		A String
		B [String]
		C Int
	 }

	 enum StringTypeEnum {
		A String
		B String
	 }

	 verb callTodoCreate(todo.CreateRequest) todo.CreateResponse
      +calls todo.create
	}
	`
	actual, err := ParseModuleString("", input)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, Normalise(testSchema.Modules[2]), actual, assert.Exclude[Position]())
}

var testSchema = MustValidate(&Schema{
	Modules: []*Module{
		{
			Name:     "todo",
			Comments: []string{"A comment"},
			Decls: []Decl{
				&Secret{
					Comments: []string{"Shhh"},
					Name:     "secretValue",
					Type:     &String{},
				},
				&Config{
					Comments: []string{"A config value"},
					Name:     "configValue",
					Type:     &String{},
				},
				&Database{
					Comments: []string{"A database"},
					Name:     "testdb",
					Type:     "postgres",
				},
				&Data{
					Name:   "CreateRequest",
					Export: true,
					Fields: []*Field{
						{Name: "name", Type: &Optional{Type: &Map{Key: &String{}, Value: &String{}}}, Metadata: []Metadata{&MetadataAlias{Kind: AliasKindJSON, Alias: "rqn"}}},
					},
				},
				&Data{
					Name:   "CreateResponse",
					Export: true,
					Fields: []*Field{
						{Name: "name", Type: &Array{Element: &String{}}, Metadata: []Metadata{&MetadataAlias{Kind: AliasKindJSON, Alias: "rsn"}}},
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
				&Data{
					Name:   "PersonalInfo",
					Export: true,
					Fields: []*Field{
						{Name: "name", Type: &Encrypted{Type: &String{}}},
						{Name: "age", Type: &Encrypted{Type: &Int{}}},
						{Name: "photo", Type: &Encrypted{Type: &Bytes{}}},
					},
				},
				&Verb{Name: "create",
					Export:   true,
					Request:  &Ref{Module: "todo", Name: "CreateRequest"},
					Response: &Ref{Module: "todo", Name: "CreateResponse"},
					Metadata: []Metadata{
						&MetadataCalls{Calls: []*Ref{{Module: "todo", Name: "destroy"}}},
						&MetadataDatabases{Calls: []*Ref{{Module: "todo", Name: "testdb"}}},
					}},
				&Verb{Name: "destroy",
					Export:   true,
					Request:  &Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []Type{&Unit{}, &Ref{Module: "todo", Name: "DestroyRequest"}, &Unit{}}},
					Response: &Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []Type{&Ref{Module: "todo", Name: "DestroyResponse"}, &String{}}},
					Metadata: []Metadata{
						&MetadataIngress{
							Type:   "http",
							Method: "GET",
							Path: []IngressPathComponent{
								&IngressPathLiteral{Text: "todo"},
								&IngressPathLiteral{Text: "destroy"},
								&IngressPathParameter{Name: "name"},
							},
						},
					},
				},
				&Verb{Name: "scheduled",
					Request:  &Unit{Unit: true},
					Response: &Unit{Unit: true},
					Metadata: []Metadata{
						&MetadataCronJob{
							Cron: "*/10 * * 1-10,11-31 * * *",
						},
					},
				},
				&Verb{Name: "twiceADay",
					Request:  &Unit{Unit: true},
					Response: &Unit{Unit: true},
					Metadata: []Metadata{
						&MetadataCronJob{
							Cron: "12h",
						},
					},
				},
				&Verb{Name: "mondays",
					Request:  &Unit{Unit: true},
					Response: &Unit{Unit: true},
					Metadata: []Metadata{
						&MetadataCronJob{
							Cron: "Mon",
						},
					},
				},
			},
		},
		{
			Name: "foo",
			Decls: []Decl{
				&Enum{
					Comments: []string{"A comment"},
					Name:     "Color",
					Type:     &String{},
					Variants: []*EnumVariant{
						{Name: "Red", Value: &StringValue{Value: "Red"}},
						{Name: "Blue", Value: &StringValue{Value: "Blue"}},
						{Name: "Green", Value: &StringValue{Value: "Green"}},
					},
				},
				&Enum{
					Name:   "ColorInt",
					Type:   &Int{},
					Export: true,
					Variants: []*EnumVariant{
						{Name: "Red", Value: &IntValue{Value: 0}},
						{Name: "Blue", Value: &IntValue{Value: 1}},
						{Name: "Green", Value: &IntValue{Value: 2}},
					},
				},
				&Enum{
					Name: "TypeEnum",
					Variants: []*EnumVariant{
						{Name: "A", Value: &TypeValue{Value: Type(&String{})}},
						{Name: "B", Value: &TypeValue{Value: Type(&Array{Element: &String{}})}},
						{Name: "C", Value: &TypeValue{Value: Type(&Int{})}},
					},
				},
				&Enum{
					Name: "StringTypeEnum",
					Variants: []*EnumVariant{
						{Name: "A", Value: &TypeValue{Value: Type(&String{})}},
						{Name: "B", Value: &TypeValue{Value: Type(&String{})}},
					},
				},
				&Verb{Name: "callTodoCreate",
					Request:  &Ref{Module: "todo", Name: "CreateRequest"},
					Response: &Ref{Module: "todo", Name: "CreateResponse"},
					Metadata: []Metadata{
						&MetadataCalls{Calls: []*Ref{{Module: "todo", Name: "create"}}},
					}},
			},
		},
		{
			Name: "payments",
			Decls: []Decl{
				&Data{Name: "OnlinePaymentCreated"},
				&Data{Name: "OnlinePaymentPaid"},
				&Data{Name: "OnlinePaymentFailed"},
				&Data{Name: "OnlinePaymentCompleted"},
				&Verb{Name: "created",
					Request:  &Ref{Module: "payments", Name: "OnlinePaymentCreated"},
					Response: &Unit{},
				},
				&Verb{Name: "paid",
					Request:  &Ref{Module: "payments", Name: "OnlinePaymentPaid"},
					Response: &Unit{},
				},
				&Verb{Name: "failed",
					Request:  &Ref{Module: "payments", Name: "OnlinePaymentFailed"},
					Response: &Unit{},
				},
				&Verb{Name: "completed",
					Request:  &Ref{Module: "payments", Name: "OnlinePaymentCompleted"},
					Response: &Unit{},
				},
				&FSM{
					Name:  "payment",
					Start: []*Ref{{Module: "payments", Name: "created"}, {Module: "payments", Name: "paid"}},
					Transitions: []*FSMTransition{
						{From: &Ref{Module: "payments", Name: "created"}, To: &Ref{Module: "payments", Name: "paid"}},
						{From: &Ref{Module: "payments", Name: "created"}, To: &Ref{Module: "payments", Name: "failed"}},
						{From: &Ref{Module: "payments", Name: "paid"}, To: &Ref{Module: "payments", Name: "completed"}},
					},
				},
			},
		},
		{
			Name: "typealias",
			Decls: []Decl{
				&TypeAlias{
					Name: "NonFtlType",
					Type: &Any{},
					Metadata: []Metadata{
						&MetadataTypeMap{Runtime: "go", NativeName: "github.com/foo/bar.Type"},
						&MetadataTypeMap{Runtime: "kotlin", NativeName: "com.foo.bar.Type"},
					},
				},
			},
		},
	},
})

func TestRetryParsing(t *testing.T) {
	for _, tt := range []struct {
		input   string
		seconds int
	}{
		{"7s", 7},
		{"9h", 9 * 60 * 60},
		{"1d", 24 * 60 * 60},
		{"1m90s", 60 + 90},
		{"1h2m3s", 60*60 + 2*60 + 3},
	} {
		duration, err := parseRetryDuration(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, time.Second*time.Duration(tt.seconds), duration)
	}
}

func TestParseTypeMap(t *testing.T) {
	input := `
	module typealias {
	 typealias NonFtlType Any
      +typemap go "github.com/foo/bar.Type"
      +typemap kotlin "com.foo.bar.Type"
	}
	`
	actual, err := ParseModuleString("", input)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, testSchema.Modules[4], actual, assert.Exclude[Position]())
}
