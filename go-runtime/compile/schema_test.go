package compile

import (
	"context"
	"fmt"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/TBD54566975/ftl/backend/schema"
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
	prebuildTestModule(t, "testdata/one", "testdata/two")

	_, actual, err := ExtractModuleSchema("testdata/one")
	assert.NoError(t, err)
	actual = schema.Normalise(actual)
	expected := `module one {
  config configValue one.Config

  data Config {
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
    enumRef two.TwoEnum
  }

  data Resp {
  }

  data SinkReq {
  }

  data SourceResp {
  }

  enum Color(String) {
    Red("Red")
    Blue("Blue")
    Green("Green")
    Yellow("Yellow")
  }

  // Comments about ColorInt.
  enum ColorInt(Int) {
    // RedInt is a color.
    RedInt(0)
    BlueInt(1)
    // GreenInt is also a color.
    GreenInt(2)
    YellowInt(3)
  }

  enum IotaExpr(Int) {
    First(1)
    Second(3)
    Third(5)
  }

  enum SimpleIota(Int) {
    Zero(0)
    One(1)
    Two(2)
  }

  secret secretValue String

  verb nothing(Unit) Unit

  verb sink(one.SinkReq) Unit

  verb source(Unit) one.SourceResp

  verb verb(one.Req) one.Resp
}
`
	assert.Equal(t, expected, actual.String())
}

func TestExtractModuleSchemaTwo(t *testing.T) {
	_, actual, err := ExtractModuleSchema("testdata/two")
	assert.NoError(t, err)
	actual = schema.Normalise(actual)
	expected := `module two {
		data Payload<T> {
		  body T
		}
	  
		data User {
		  name String
		}
	  
		data UserResponse {
		  user two.User
		}
	  
		enum TwoEnum(String) {
		  Red("Red")
		  Blue("Blue")
		  Green("Green")
		}
	  
		verb callsTwo(two.Payload<String>) two.Payload<String>  
			+calls two.two
	  
		verb returnsUser(Unit) two.UserResponse?
	  
		verb two(two.Payload<String>) two.Payload<String>
	  }
`
	fmt.Printf("actual: %s\n", actual.String())
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func TestParseDirectives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected directive
	}{
		{name: "Module", input: "ftl:module foo", expected: &directiveModule{Name: "foo"}},
		{name: "Verb", input: "ftl:verb", expected: &directiveVerb{Verb: true}},
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := directiveParser.ParseString("", tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got.Directive, assert.Exclude[lexer.Position](), assert.Exclude[schema.Position]())
		})
	}
}

func TestParseTypesTime(t *testing.T) {
	timeRef := mustLoadRef("time", "Time").Type()
	parsed, err := visitType(nil, token.NoPos, timeRef)
	assert.NoError(t, err)
	_, ok := parsed.(*schema.Time)
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
			parsed, err := visitType(nil, token.NoPos, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}

func normaliseString(s string) string {
	return strings.TrimSpace(strings.Join(slices.Map(strings.Split(s, "\n"), strings.TrimSpace), "\n"))
}

func TestErrorReporting(t *testing.T) {
	pwd, _ := os.Getwd()
	_, _, err := ExtractModuleSchema("testdata/failing")
	assert.EqualError(t, err, filepath.Join(pwd, `testdata/failing/failing.go`)+`:15:2: call must have exactly three arguments`)
}
