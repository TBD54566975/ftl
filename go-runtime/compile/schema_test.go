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

	"github.com/TBD54566975/golang-tools/go/packages"

	"github.com/TBD54566975/ftl/backend/schema"
	extract "github.com/TBD54566975/ftl/go-runtime/schema"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

// this is helpful when a test requires another module to be built before running
// eg: when module A depends on module B, we need to build module B before building module A
func prebuildTestModule(t *testing.T, args ...string) error {
	t.Helper()

	ctx := log.ContextWithNewDefaultLogger(context.Background())

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	ftlArgs := []string{"build"}
	ftlArgs = append(ftlArgs, args...)

	cmd := exec.Command(ctx, log.Debug, dir, "ftl", ftlArgs...)
	err = cmd.RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("ftl build failed: %w", err)
	}
	return nil
}

func TestExtractModuleSchema(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	assert.NoError(t, prebuildTestModule(t, "testdata/one", "testdata/two"))

	r, err := ExtractModuleSchema("testdata/one", &schema.Schema{})
	assert.NoError(t, err)
	actual := schema.Normalise(r.Module)
	expected := `module one {
  config configValue one.Config 
  secret secretValue String

  database postgres testDb

  export enum BlobOrList {
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

  enum PrivateEnum {
    ExportedStruct one.ExportedStruct
    PrivateStruct one.PrivateStruct
    WithoutDirectiveStruct one.WithoutDirectiveStruct
  }

  enum SimpleIota: Int {
    Zero = 0
    One = 1
    Two = 2
  }

  enum TypeEnum {
    Option String?
    InlineStruct one.InlineStruct
    AliasedStruct one.UnderlyingStruct
    ValueEnum one.ColorInt
  }

  data Config {
    field String
  }

  data DataWithType<T> {
    value T
  }

  export data ExportedData {
    field String
  }

  export data ExportedStruct {
  }

  data InlineStruct {
  }

  export data Nested {
  }

  data PrivateStruct {
  }

  export data Req {
    int Int
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

  export data Resp {
  }

  data SinkReq {
  }

  data SourceResp {
  }

  data UnderlyingStruct {
  }

  data WithoutDirectiveStruct {
  }

  export verb http(builtin.HttpRequest<one.Req>) builtin.HttpResponse<one.Resp, Unit>
    +ingress http GET /get

  export verb nothing(Unit) Unit

  verb sink(one.SinkReq) Unit

  verb source(Unit) one.SourceResp

  verb verb(one.Req) one.Resp
}
`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func TestExtractModuleSchemaTwo(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	assert.NoError(t, prebuildTestModule(t, "testdata/two"))

	r, err := ExtractModuleSchema("testdata/two", &schema.Schema{})
	assert.NoError(t, err)
	assert.Equal(t, r.Errors, nil)
	actual := schema.Normalise(r.Module)
	expected := `module two {
		export enum TwoEnum: String {
		  Red = "Red"
		  Blue = "Blue"
		  Green = "Green"
        }

        export enum TypeEnum {
		  Scalar String
		  List [String]
		  Exported two.Exported
		  WithoutDirective two.WithoutDirective
		}

		export data Exported {
		}

		export data Payload<T> {
		  body T
		}

		export data User {
		  name String
		}

		export data UserResponse {
		  user two.User
		}

		export data WithoutDirective {
		}

		export verb callsTwo(two.Payload<String>) two.Payload<String>
			+calls two.two

		export verb returnsUser(Unit) two.UserResponse

		export verb two(two.Payload<String>) two.Payload<String>
	  }
`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func TestExtractModuleSchemaFSM(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	r, err := ExtractModuleSchema("testdata/fsm", &schema.Schema{})
	assert.NoError(t, err)
	assert.Equal(t, nil, r.Errors, "expected no schema errors")
	actual := schema.Normalise(r.Module)
	expected := `module fsm {
		fsm payment
			+retry 10 5s 10m
		{
			start fsm.created
			start fsm.paid
			transition fsm.created to fsm.paid
			transition fsm.created to fsm.failed
			transition fsm.paid to fsm.completed
		}

		// The message to be sent when the payment is completed.
		//
		// Otherwise, OnlinePaymentFailed will be sent.
		data OnlinePaymentCompleted {
		}

		data OnlinePaymentCreated {
		}

		data OnlinePaymentFailed {
		}

		data OnlinePaymentPaid {
		}

		verb completed(fsm.OnlinePaymentCompleted) Unit
			+retry 1s

		verb created(fsm.OnlinePaymentCreated) Unit
			+retry 5 1m30s 7m

		verb failed(fsm.OnlinePaymentFailed) Unit
			+retry 5 1h 1d

		// The message to be sent when the payment is paid.
		verb paid(fsm.OnlinePaymentPaid) Unit
			+retry 5 60s
	}
`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func TestExtractModuleSchemaNamedTypes(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	assert.NoError(t, prebuildTestModule(t, "testdata/named", "testdata/namedext"))
	r, err := ExtractModuleSchema("testdata/named", &schema.Schema{})
	assert.NoError(t, err)
	assert.Equal(t, nil, r.Errors, "expected no schema errors")
	actual := schema.Normalise(r.Module)
	expected := `module named {
		typealias DoubleAliasedUser named.InternalUser

		// ID testing if typealias before struct works
		export typealias Id String

		typealias InternalUser named.User

		// Name testing if typealias after struct works
		export typealias Name String

		// UserSource, testing that defining an enum after struct works
		export enum UserSource: String {
			Magazine = "magazine"
			Friend = "friend"
			Ad = "ad"
		}

		// UserState, testing that defining an enum before struct works
		export enum UserState: String {
			Onboarded = "onboarded"
			Registered = "registered"
			Active = "active"
			Inactive = "inactive"
		}

		export data User {
			id named.Id
			name named.Name
			state named.UserState
			source named.UserSource
			comment namedext.Comment
			emailConsent namedext.EmailConsent
		}

		verb pingInternalUser(named.InternalUser) Unit

		verb pingUser(named.User) Unit
	}
	`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func TestExtractModuleSchemaParent(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	assert.NoError(t, prebuildTestModule(t, "testdata/parent"))
	r, err := ExtractModuleSchema("testdata/parent", &schema.Schema{})
	assert.NoError(t, err)
	assert.Equal(t, nil, r.Errors, "expected no schema errors")
	actual := schema.Normalise(r.Module)
	expected := `module parent {
		export typealias ChildAlias String

		export data ChildStruct {
			name parent.ChildAlias?
		}

		data Resp {
		}

		verb childVerb(Unit) parent.Resp

		export verb verb(Unit) parent.ChildStruct
	}
	`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func TestExtractModulePubSub(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	assert.NoError(t, prebuildTestModule(t, "testdata/pubsub"))

	r, err := ExtractModuleSchema("testdata/pubsub", &schema.Schema{})
	assert.NoError(t, err)
	assert.Equal(t, nil, r.Errors, "expected no schema errors")
	actual := schema.Normalise(r.Module)
	expected := `module pubsub {
		topic payins pubsub.PayinEvent
		// publicBroadcast is a topic that broadcasts payin events to the public.
		// out of order with subscription registration to test ordering doesn't matter.
		export topic publicBroadcast pubsub.PayinEvent
		subscription broadcastSubscription pubsub.publicBroadcast
		subscription paymentProcessing pubsub.payins

        export data PayinEvent {
        	name String
        }

		export verb broadcast(Unit) Unit

        verb payin(Unit) Unit

        verb processBroadcast(pubsub.PayinEvent) Unit
        	+subscribe broadcastSubscription
			+retry 10 1s

        verb processPayin(pubsub.PayinEvent) Unit
        	+subscribe paymentProcessing
	}
`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func TestExtractModuleSubscriber(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	assert.NoError(t, prebuildTestModule(t, "testdata/pubsub", "testdata/subscriber"))
	r, err := ExtractModuleSchema("testdata/subscriber", &schema.Schema{})
	assert.NoError(t, err)
	assert.Equal(t, nil, r.Errors, "expected no schema errors")
	actual := schema.Normalise(r.Module)
	expected := `module subscriber {
		subscription subscriptionToExternalTopic pubsub.publicBroadcast

        verb consumesSubscriptionFromExternalTopic(pubsub.PayinEvent) Unit
		+subscribe subscriptionToExternalTopic
	}
`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func TestParsedirectives(t *testing.T) {
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
		{name: "TypeAlias", input: "ftl:typealias", expected: &directiveTypeAlias{TypeAlias: true}},
		{name: "TypeAlias export", input: "ftl:typealias export", expected: &directiveTypeAlias{TypeAlias: true, Export: true}},
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
	pctx := newParseContext(nil, []*packages.Package{}, &schema.Schema{}, &extract.Result{Module: &schema.Module{}})
	parsed, ok := visitType(pctx, token.NoPos, timeRef, false).Get()
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

	_ = prebuildTestModule(t, "testdata/failing", "testdata/pubsub") //nolint:errcheck // prebuild so we have external_module.go for pubsub module, but ignore these initial errors

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	err = exec.Command(ctx, log.Debug, "testdata/failing", "go", "mod", "tidy").RunBuffered(ctx)
	assert.NoError(t, err)
	r, err := ExtractModuleSchema("testdata/failing", &schema.Schema{})
	assert.NoError(t, err)

	filename := filepath.Join(pwd, `testdata/failing/failing.go`)
	subFilename := filepath.Join(pwd, `testdata/failing/child/child.go`)
	actual := slices.Map(r.Errors, func(e *schema.Error) string {
		str := strings.ReplaceAll(e.Error(), filename+":", "")
		str = strings.ReplaceAll(str, subFilename+":", "")
		return str
	})
	expected := []string{
		// failing/child/child.go
		`4:2-6: unsupported type "uint64" for field "Body"`,

		// failing/failing.go
		`13:13-34: expected string literal for argument at index 0`,
		`16:18-52: duplicate config declaration at 15:18-52`,
		`19:18-52: duplicate secret declaration at 18:18-52`,
		`22:14-44: duplicate database declaration at 21:14-44`,
		`25:2-10: unsupported type "error" for field "BadParam"`,
		`28:2-17: unsupported type "uint64" for field "AnotherBadParam"`,
		`31:3-3: unexpected directive "ftl:export" attached for verb, did you mean to use '//ftl:verb export' instead?`,
		`37:36-36: unsupported request type "ftl/failing.Request"`,
		`37:50-50: unsupported response type "ftl/failing.Response"`,
		`38:16-29: call first argument must be a function but is an unresolved reference to lib.OtherFunc`,
		`38:16-29: call first argument must be a function in an ftl module`,
		`39:2-46: call must have exactly three arguments`,
		`40:16-25: call first argument must be a function in an ftl module`,
		`45:1-2: must have at most two parameters (context.Context, struct)`,
		`45:69-69: unsupported response type "ftl/failing.Response"`,
		`50:22-27: first parameter must be of type context.Context but is ftl/failing.Request`,
		`50:37-43: second parameter must be a struct but is string`,
		`50:53-53: unsupported response type "ftl/failing.Response"`,
		`55:43-47: second parameter must not be ftl.Unit`,
		`55:59-59: unsupported response type "ftl/failing.Response"`,
		`60:1-2: first parameter must be context.Context`,
		`60:18-18: unsupported response type "ftl/failing.Response"`,
		`65:1-2: must have at most two results (<type>, error)`,
		`65:41-41: unsupported request type "ftl/failing.Request"`,
		`70:1-2: must at least return an error`,
		`70:36-36: unsupported request type "ftl/failing.Request"`,
		`74:35-35: unsupported request type "ftl/failing.Request"`,
		`74:48-48: must return an error but is ftl/failing.Response`,
		`79:41-41: unsupported request type "ftl/failing.Request"`,
		`79:55-55: first result must be a struct but is string`,
		`79:63-63: must return an error but is string`,
		`79:63-63: second result must not be ftl.Unit`,
		// `86:1-2: duplicate declaration of "WrongResponse" at 79:6`,  TODO: fix this
		`90:3-3: unexpected directive "ftl:verb"`,
		`104:2-24: cannot attach enum value to BadValueEnum because it is a variant of type enum TypeEnum, not a value enum`,
		`111:2-41: cannot attach enum value to BadValueEnumOrderDoesntMatter because it is a variant of type enum TypeEnum, not a value enum`,
		`124:21-60: config and secret names must be valid identifiers`,
		`130:1-1: schema declaration contains conflicting directives`,
		`130:1-26: only one directive expected when directive "ftl:enum" is present, found multiple`,
		`130:1-26: only one directive expected when directive "ftl:typealias" is present, found multiple`,
		`146:1-35: type can not be a variant of more than 1 type enums (TypeEnum1, TypeEnum2)`,
		`152:27-27: enum discriminator "TypeEnum3" cannot contain exported methods`,
		`155:1-35: enum discriminator "NoMethodsTypeEnum" must define at least one method`,
		`167:3-14: unexpected token "d"`,
		`174:2-62: can not publish directly to topics in other modules`,
		`175:9-26: can not call verbs in other modules directly: use ftl.Call(…) instead`,
		`180:2-12: struct field unexported must be exported by starting with an uppercase letter`,
		`184:6-6: unsupported type "ftl/failing/child.BadChildStruct" for field "child"`,
	}
	assert.Equal(t, expected, actual)
}

// Where parsing is correct but validation of the schema fails
func TestValidationFailures(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	err = exec.Command(ctx, log.Debug, "testdata/validation", "go", "mod", "tidy").RunBuffered(ctx)
	assert.NoError(t, err)
	_, err = ExtractModuleSchema("testdata/validation", &schema.Schema{})
	assert.Error(t, err)
	errs := errors.UnwrapAll(err)

	filename := filepath.Join(pwd, `testdata/validation/validation.go`)
	actual := slices.Map(errs, func(e error) string {
		return strings.TrimPrefix(e.Error(), filename+":")
	})
	expected := []string{
		`11:3-3: verb badYear: invalid cron expression "* * * * * 9999": value 9999 out of allowed year range of 0-3000`,
		`16:3-3: verb allZeroes: invalid cron expression "0 0 0 0 0": value 0 out of allowed day of month range of 1-31`,
	}
	assert.Equal(t, expected, actual)
}
