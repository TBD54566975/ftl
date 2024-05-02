package compile

import (
	"context"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

// this is helpful when a test requires another module to be built before running
// eg: when module A depends on module B, we need to build module B before building module A
func prebuildTestModule(t *testing.T, args ...string) {
	t.Helper()

	ctx := log.ContextWithNewDefaultLogger(context.Background())

	dir, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current directory: %v", err)

	ftlArgs := []string{"build"}
	ftlArgs = append(ftlArgs, args...)

	cmd := exec.Command(ctx, log.Debug, dir, "ftl", ftlArgs...)
	err = cmd.RunBuffered(ctx)
	assert.NoError(t, err, "ftl build failed with %s\n", err)
}

func TestExtractModuleSchema(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	prebuildTestModule(t, "testdata/one", "testdata/two")

	_, actual, _, err := ExtractModuleSchema("testdata/one")
	assert.NoError(t, err)
	actual = schema.Normalise(actual)
	expected := `module one {
  config configValue one.Config
  secret secretValue String

  postgres database testDb

  enum BlobOrList {
    Blob String
    List [String]
  }

  export enum Color: String {
    Red = "Red"
    Blue = "Blue"
    Green = "Green"
    Yellow = "Yellow"
  }

  // Comments about ColorInt.
  enum ColorInt: Int {
    // RedInt is a color.
    RedInt = 0
    BlueInt = 1
    // GreenInt is also a color.
    GreenInt = 2
    YellowInt = 3
  }

  enum IotaExpr: Int {
    First = 1
    Second = 3
    Third = 5
  }

  enum SimpleIota: Int {
    Zero = 0
    One = 1
    Two = 2
  }

  data Config {
    field String
  }

  export data ExportedData {
    field String
  }

  data Nested {
  }

  data Req {
    int Int
    int64 Int
    float Float
    string String
    slice [String]
    map {String: String}
    nested one.Nested
    optional one.Nested?
    time Time
    user two.User +alias json "u"
    bytes Bytes
    localValueEnumRef one.Color
    localTypeEnumRef one.BlobOrList
    externalValueEnumRef two.TwoEnum
    externalTypeEnumRef two.TypeEnum
  }

  data Resp {
  }

  data SinkReq {
  }

  data SourceResp {
  }

  export verb http(builtin.HttpRequest<one.Req>) builtin.HttpResponse<one.Resp, Unit>  
      +ingress http GET /get

  export verb nothing(Unit) Unit

  verb sink(one.SinkReq) Unit

  verb source(Unit) one.SourceResp

  verb verb(one.Req) one.Resp
}
`
	assert.Equal(t, expected, actual.String())
}

func TestExtractModuleSchemaTwo(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	_, actual, _, err := ExtractModuleSchema("testdata/two")
	assert.NoError(t, err)
	actual = schema.Normalise(actual)
	expected := `module two {
		export enum TwoEnum: String {
		  Red = "Red"
		  Blue = "Blue"
		  Green = "Green"
        }

                enum TypeEnum {
		  Scalar String
		  List [String]
		  Exported two.Exported
		  Private two.Private
		  WithoutDirective two.WithoutDirective
		}

		export data Exported {
		}

		export data Payload<T> {
		  body T
		}

		data Private {
		}

		export data User {
		  name String
		}

		export data UserResponse {
		  user two.User
		}

		data WithoutDirective {
		}

		export verb callsTwo(two.Payload<String>) two.Payload<String>
			+calls two.two

		export verb returnsUser(Unit) two.UserResponse

		export verb two(two.Payload<String>) two.Payload<String>
	  }
`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func TestParseDirectives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected directive
	}{
		{name: "Verb", input: "ftl:verb", expected: &directiveVerb{Verb: true}},
		{name: "Verb export", input: "ftl:verb export", expected: &directiveVerb{Verb: true, Export: true}},
		{name: "Data", input: "ftl:data", expected: &directiveData{Data: true}},
		{name: "Data export", input: "ftl:data export", expected: &directiveData{Data: true, Export: true}},
		{name: "Enum", input: "ftl:enum", expected: &directiveEnum{Enum: true}},
		{name: "Enum export", input: "ftl:enum export", expected: &directiveEnum{Enum: true, Export: true}},
		{name: "Ingress", input: `ftl:ingress GET /foo`, expected: &directiveIngress{
			Method: "GET",
			Path: []schema.IngressPathComponent{
				&schema.IngressPathLiteral{
					Text: "foo",
				},
			},
		}},
		{name: "Ingress", input: `ftl:ingress GET /test_path/{something}/987-Your_File.txt%7E%21Misc%2A%28path%29info%40abc%3Fxyz`, expected: &directiveIngress{
			Method: "GET",
			Path: []schema.IngressPathComponent{
				&schema.IngressPathLiteral{
					Text: "test_path",
				},
				&schema.IngressPathParameter{
					Name: "something",
				},
				&schema.IngressPathLiteral{
					Text: "987-Your_File.txt%7E%21Misc%2A%28path%29info%40abc%3Fxyz",
				},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := directiveParser.ParseString("", tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got.Directive, assert.Exclude[lexer.Position](), assert.Exclude[schema.Position]())
		})
	}
}

func TestParseTypesTime(t *testing.T) {
	timeRef := mustLoadRef("time", "Time").Type()
	parsed, ok := visitType(nil, token.NoPos, timeRef, false).Get()
	assert.True(t, ok)
	_, ok = parsed.(*schema.Time)
	assert.True(t, ok)
}

func TestParseBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Type
		expected schema.Type
	}{
		{name: "String", input: types.Typ[types.String], expected: &schema.String{}},
		{name: "Int", input: types.Typ[types.Int], expected: &schema.Int{}},
		{name: "Bool", input: types.Typ[types.Bool], expected: &schema.Bool{}},
		{name: "Float64", input: types.Typ[types.Float64], expected: &schema.Float{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, ok := visitType(nil, token.NoPos, tt.input, false).Get()
			assert.True(t, ok)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}

func normaliseString(s string) string {
	return strings.TrimSpace(strings.Join(slices.Map(strings.Split(s, "\n"), strings.TrimSpace), "\n"))
}

func TestErrorReporting(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	pwd, _ := os.Getwd()
	_, _, schemaErrs, err := ExtractModuleSchema("testdata/failing")
	assert.NoError(t, err)

	filename := filepath.Join(pwd, `testdata/failing/failing.go`)
	assert.EqualError(t, errors.Join(genericizeErrors(schemaErrs)...),
		filename+":10:13-35: config and secret declarations must have a single string literal argument\n"+
			filename+":13:18-52: duplicate config declaration at 12:18-52\n"+
			filename+":16:18-52: duplicate secret declaration at 15:18-52\n"+
			filename+":19:14-44: duplicate database declaration at 18:14-44\n"+
			filename+":22:2-10: unsupported type \"error\" for field \"BadParam\"\n"+
			filename+":25:2-17: unsupported type \"uint64\" for field \"AnotherBadParam\"\n"+
			filename+":28:3-3: unexpected token \"export\" (expected Directive)\n"+
			filename+":34:36-39: unsupported request type \"ftl/failing.Request\"\n"+
			filename+":34:50-50: unsupported response type \"ftl/failing.Response\"\n"+
			filename+":35:16-29: call first argument must be a function in an ftl module\n"+
			filename+":36:2-46: call must have exactly three arguments\n"+
			filename+":37:16-25: call first argument must be a function\n"+
			filename+":42:1-2: must have at most two parameters (context.Context, struct)\n"+
			filename+":42:69-69: unsupported response type \"ftl/failing.Response\"\n"+
			filename+":47:22-27: first parameter must be of type context.Context but is ftl/failing.Request\n"+
			filename+":47:37-43: second parameter must be a struct but is string\n"+
			filename+":47:53-53: unsupported response type \"ftl/failing.Response\"\n"+
			filename+":52:43-47: second parameter must not be ftl.Unit\n"+
			filename+":52:59-59: unsupported response type \"ftl/failing.Response\"\n"+
			filename+":57:1-2: first parameter must be context.Context\n"+
			filename+":57:18-18: unsupported response type \"ftl/failing.Response\"\n"+
			filename+":62:1-2: must have at most two results (struct, error)\n"+
			filename+":62:41-44: unsupported request type \"ftl/failing.Request\"\n"+
			filename+":67:1-2: must at least return an error\n"+
			filename+":67:36-39: unsupported request type \"ftl/failing.Request\"\n"+
			filename+":71:35-38: unsupported request type \"ftl/failing.Request\"\n"+
			filename+":71:48-48: must return an error but is ftl/failing.Response\n"+
			filename+":76:41-44: unsupported request type \"ftl/failing.Request\"\n"+
			filename+":76:55-55: first result must be a struct but is string\n"+
			filename+":76:63-63: must return an error but is string\n"+
			filename+":76:63-63: second result must not be ftl.Unit\n"+
			filename+":83:1-1: duplicate verb name \"WrongResponse\"\n"+
			filename+":89:2-12: struct field unexported must be exported by starting with an uppercase letter\n"+
			filename+":100:2-23: cannot attach enum value to BadValueEnum because it is a variant of type enum TypeEnum, not a value enum\n"+
			filename+":106:2-40: cannot attach enum value to BadValueEnumOrderDoesntMatter because it is a variant of type enum TypeEnum, not a value enum",
	)
}

func genericizeErrors(schemaErrs []*schema.Error) []error {
	errs := make([]error, len(schemaErrs))
	for i, schemaErr := range schemaErrs {
		errs[i] = schemaErr
	}
	return errs
}
