//go:build integration

package schema

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/TBD54566975/ftl/go-runtime/schema/common"

	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"
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

func TestExtractSchema(t *testing.T) {
	t.Run("ExtractModuleSchema", testExtractModuleSchema)
	t.Run("TestExtractModuleSchemaTwo", testExtractModuleSchemaTwo)
	t.Run("TestExtractModuleSchemaNamedTypes", testExtractModuleSchemaNamedTypes)
	t.Run("TestExtractModuleSchemaParent", testExtractModuleSchemaParent)
	t.Run("TestExtractModulePubSub", testExtractModulePubSub)
	t.Run("TestExtractModuleSubscriber", testExtractModuleSubscriber)
	t.Run("TestParsedirectives", testParsedirectives)
	t.Run("TestErrorReporting", testErrorReporting)
	t.Run("TestValidationFailures", testValidationFailures)
}
func testExtractModuleSchema(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	assert.NoError(t, prebuildTestModule(t, "testdata/one", "testdata/two"))

	r, err := Extract("testdata/one")
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
    Blue = "Blue"
    Green = "Green"
    Red = "Red"
    Yellow = "Yellow"
  }

  // Comments about ColorInt.
  enum ColorInt: Int {
    BlueInt = 1
    // GreenInt is also a color.
    GreenInt = 2
    // RedInt is a color.
    RedInt = 0
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
    One = 1
    Two = 2
    Zero = 0
  }

  enum TypeEnum {
    AliasedStruct one.UnderlyingStruct
    InlineStruct one.InlineStruct
    Option String?
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

  verb batchStringToTime([String]) [Time]
  
  export verb http(builtin.HttpRequest<Unit, Unit, one.Req>) builtin.HttpResponse<one.Resp, Unit>
    +ingress http GET /get

  export verb nothing(Unit) Unit

  verb sink(one.SinkReq) Unit

  verb source(Unit) one.SourceResp

  verb stringToTime(String) Time

  verb verb(one.Req) one.Resp
}
`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func testExtractModuleSchemaTwo(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	assert.NoError(t, prebuildTestModule(t, "testdata/two"))

	r, err := Extract("testdata/two")
	assert.NoError(t, err)
	for _, e := range r.Errors {
		// only warns
		assert.True(t, e.Level == builderrors.WARN)
	}
	actual := schema.Normalise(r.Module)
	expected := `module two {
			typealias BackoffAlias Any
        		+typemap go "github.com/jpillora/backoff.Backoff"
			typealias ExplicitAliasAlias Any
				+typemap kotlin "com.foo.bar.NonFTLType"
				+typemap go "github.com/TBD54566975/ftl/go-runtime/schema/testdata.lib.NonFTLType"
			typealias ExplicitAliasType Any
				+typemap kotlin "com.foo.bar.NonFTLType"
				+typemap go "github.com/TBD54566975/ftl/go-runtime/schema/testdata.lib.NonFTLType"
			typealias PaymentState String
			typealias TransitiveAliasAlias Any
				+typemap go "github.com/TBD54566975/ftl/go-runtime/schema/testdata.lib.NonFTLType"
			typealias TransitiveAliasType Any
				+typemap go "github.com/TBD54566975/ftl/go-runtime/schema/testdata.lib.NonFTLType"

			enum PayinState: two.PaymentState {
			  PayinPending = "PAYIN_PENDING"
			}

			enum PayoutState: two.PaymentState {
			  PayoutPending = "PAYOUT_PENDING"
			}

			export enum TwoEnum: String {
			  Blue = "Blue"
			  Green = "Green"
			  Red = "Red"
	        }

	        export enum TypeEnum {
			  Exported two.Exported
			  List [String]
			  Scalar String
			  WithoutDirective two.WithoutDirective
			}

			export data Exported {
			}

			data NonFtlField {
			  explicitType two.ExplicitAliasType
			  explicitAlias two.ExplicitAliasAlias
			  transitiveType two.TransitiveAliasType
			  transitiveAlias two.TransitiveAliasAlias
			}

			export data Payload<T> {
			  body T
			}

			data Payment {
			  in two.PayinState
			  out two.PayoutState
			}

			export data PostRequest {
			  userId Int
			  postId Int
			}

			export data PostResponse {
			  success Bool
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

			export verb callsTwoAndThree(two.Payload<String>) two.Payload<String>
				+calls two.two, two.three

			export verb ingress(builtin.HttpRequest<two.PostRequest, Unit, Unit>) builtin.HttpResponse<two.PostResponse, String>
				+ingress http POST /users
				+encoding json lenient

			export verb returnsUser(Unit) two.UserResponse

			export verb three(two.Payload<String>) two.Payload<String>

			export verb two(two.Payload<String>) two.Payload<String>
		  }
	`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func testExtractModuleSchemaNamedTypes(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	assert.NoError(t, prebuildTestModule(t, "testdata/named", "testdata/namedext"))
	r, err := Extract("testdata/named")
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
			Ad = "ad"
			Friend = "friend"
			Magazine = "magazine"
		}

		// UserState, testing that defining an enum before struct works
		export enum UserState: String {
			Active = "active"
			Inactive = "inactive"
			// Out of order
			Onboarded = "onboarded"
			Registered = "registered"
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

func testExtractModuleSchemaParent(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	assert.NoError(t, prebuildTestModule(t, "testdata/parent"))
	r, err := Extract("testdata/parent")
	assert.NoError(t, err)
	assert.Equal(t, nil, r.Errors, "expected no schema errors")
	actual := schema.Normalise(r.Module)
	expected := `module parent {
		export typealias ChildAlias String

		export enum ChildTypeEnum {
			List [String]
			Scalar String
		}

		export enum ChildValueEnum: Int {
			A = 0
			B = 1
			C = 2
		}

		export data ChildStruct {
			name parent.ChildAlias?
			valueEnum parent.ChildValueEnum
			typeEnum parent.ChildTypeEnum
		}

		data Resp {
		}

		verb childVerb(Unit) parent.Resp

		export verb verb(Unit) parent.ChildStruct
	}
	`
	assert.Equal(t, normaliseString(expected), normaliseString(actual.String()))
}

func testExtractModulePubSub(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	assert.NoError(t, prebuildTestModule(t, "testdata/pubsub"))

	r, err := Extract("testdata/pubsub")
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

func testExtractModuleSubscriber(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	assert.NoError(t, prebuildTestModule(t, "testdata/pubsub", "testdata/subscriber"))
	r, err := Extract("testdata/subscriber")
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

func testParsedirectives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected common.Directive
	}{
		{name: "Verb", input: "ftl:verb", expected: &common.DirectiveVerb{Verb: true}},
		{name: "Verb export", input: "ftl:verb export", expected: &common.DirectiveVerb{Verb: true, Export: true}},
		{name: "Data", input: "ftl:data", expected: &common.DirectiveData{Data: true}},
		{name: "Data export", input: "ftl:data export", expected: &common.DirectiveData{Data: true, Export: true}},
		{name: "Enum", input: "ftl:enum", expected: &common.DirectiveEnum{Enum: true}},
		{name: "Enum export", input: "ftl:enum export", expected: &common.DirectiveEnum{Enum: true, Export: true}},
		{name: "TypeAlias", input: "ftl:typealias", expected: &common.DirectiveTypeAlias{TypeAlias: true}},
		{name: "TypeAlias export", input: "ftl:typealias export", expected: &common.DirectiveTypeAlias{TypeAlias: true, Export: true}},
		{name: "Ingress", input: `ftl:ingress GET /foo`, expected: &common.DirectiveIngress{
			Method: "GET",
			Path: []schema.IngressPathComponent{
				&schema.IngressPathLiteral{
					Text: "foo",
				},
			},
		}},
		{name: "Ingress", input: `ftl:ingress GET /test_path/{something}/987-Your_File.txt%7E%21Misc%2A%28path%29info%40abc%3Fxyz`, expected: &common.DirectiveIngress{
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
			got, err := common.DirectiveParser.ParseString("", tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got.Directive, assert.Exclude[lexer.Position](), assert.Exclude[schema.Position]())
		})
	}
}

func testErrorReporting(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	_ = prebuildTestModule(t, "testdata/failing", "testdata/pubsub") //nolint:errcheck // prebuild so we have external_module.go for pubsub module, but ignore these initial errors

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	err = exec.Command(ctx, log.Debug, "testdata/failing", "go", "mod", "tidy").RunBuffered(ctx)
	assert.NoError(t, err)
	r, err := Extract("testdata/failing")
	assert.NoError(t, err)

	var actualParent []string
	var actualChild []string
	filename := filepath.Join(pwd, `testdata/failing/failing.go`)
	subFilename := filepath.Join(pwd, `testdata/failing/child/child.go`)
	for _, e := range r.Errors {
		str := strings.ReplaceAll(e.Error(), subFilename+":", "")
		str = strings.ReplaceAll(str, filename+":", "")
		if pos, ok := e.Pos.Get(); ok && pos.Filename == filename {
			actualParent = append(actualParent, str)
		} else {
			actualChild = append(actualChild, str)
		}
	}

	// failing/failing.go
	expectedParent := []string{
		`12:13-34: expected string literal for argument at index 0`,
		`15:18: duplicate config declaration for "failing.FTL_CONFIG_ENDPOINT"; already declared at "37:18"`,
		`18:18: duplicate secret declaration for "failing.FTL_SECRET_ENDPOINT"; already declared at "38:18"`,
		`21:14: duplicate database declaration for "failing.testDb"; already declared at "41:14"`,
		`24:2-10: unsupported type "error" for field "BadParam"`,
		`27:2-17: unsupported type "uint64" for field "AnotherBadParam"`,
		`30:3: unexpected directive "ftl:export" attached for verb, did you mean to use '//ftl:verb export' instead?`,
		`36:45: unsupported request type "ftl/failing.Request"`,
		`36:54-66: unsupported verb parameter type "Request"; verbs must have the signature func(Context, Request?, Resources...)`,
		`36:69: unsupported response type "ftl/failing.Response"`,
		`41:22-27: first parameter must be of type context.Context but is ftl/failing.Request`,
		`41:53: unsupported response type "ftl/failing.Response"`,
		`46:43-47: second parameter must not be ftl.Unit`,
		`46:59: unsupported response type "ftl/failing.Response"`,
		`51:1-2: first parameter must be context.Context`,
		`51:18: unsupported response type "ftl/failing.Response"`,
		`56:1-2: must have at most two results (<type>, error)`,
		`56:45: unsupported request type "ftl/failing.Request"`,
		`61:1-2: must at least return an error`,
		`61:40: unsupported request type "ftl/failing.Request"`,
		`65:39: unsupported request type "ftl/failing.Request"`,
		`65:48: must return an error but is ftl/failing.Response`,
		`70:45: unsupported request type "ftl/failing.Request"`,
		`70:63: must return an error but is string`,
		`70:63: second result must not be ftl.Unit`,
		`81:3: unexpected directive "ftl:verb"`,
		`90:6-18: "BadValueEnum" is a value enum and cannot be tagged as a variant of type enum "TypeEnum" directly`,
		`99:6-35: "BadValueEnumOrderDoesntMatter" is a value enum and cannot be tagged as a variant of type enum "TypeEnum" directly`,
		`111:6: schema declaration with name "PrivateData" already exists for module "failing"; previously declared at "40:25"`,
		`115:21-60: config names must be valid identifiers`,
		`121:1: schema declaration contains conflicting directives`,
		`121:1-26: only one directive expected when directive "ftl:enum" is present, found multiple`,
		`143:6-45: enum discriminator "TypeEnum3" cannot contain exported methods`,
		`146:6-35: enum discriminator "NoMethodsTypeEnum" must define at least one method`,
		`158:3-14: unexpected token "d"`,
		`165:2-62: can not publish directly to topics in other modules`,
		`171:2-12: struct field unexported must be exported by starting with an uppercase letter`,
		`175:6: unsupported type "ftl/failing/child.BadChildStruct" for field "child"`,
		`180:6: duplicate data declaration for "failing.Redeclared"; already declared at "27:6"`,
		`197:9: direct verb calls are not allowed; use the provided EmptyClient instead. See https://tbd54566975.github.io/ftl/docs/reference/verbs/#calling-verbs`,
	}

	// failing/child/child.go
	expectedChild := []string{
		`9:2-6: unsupported type "uint64" for field "Body"`,
		`14:2-7: unsupported type "github.com/TBD54566975/ftl/go-runtime/schema/testdata.NonFTLType" for field "Field"`,
		`14:8: unsupported external type "github.com/TBD54566975/ftl/go-runtime/schema/testdata.NonFTLType"; see FTL docs on using external types: tbd54566975.github.io/ftl/docs/reference/externaltypes/`,
		`19:6-41: declared type github.com/blah.lib.NonFTLType in typemap does not match native type github.com/TBD54566975/ftl/go-runtime/schema/testdata.lib.NonFTLType`,
		`24:6: multiple Go type mappings found for "ftl/failing/child.MultipleMappings"`,
		`34:2-13: enum variant "SameVariant" conflicts with existing enum variant of "EnumVariantConflictParent" at "187:2"`,
	}
	assert.Equal(t, expectedParent, actualParent)
	assert.Equal(t, expectedChild, actualChild)
}

// Where parsing is correct but validation of the schema fails
func testValidationFailures(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	err = exec.Command(ctx, log.Debug, "testdata/validation", "go", "mod", "tidy").RunBuffered(ctx)
	assert.NoError(t, err)
	_, err = Extract("testdata/validation")
	assert.Error(t, err)
	errs := errors.UnwrapAll(err)

	filename := filepath.Join(pwd, `testdata/validation/validation.go`)
	actual := slices.Map(errs, func(e error) string {
		return strings.TrimPrefix(e.Error(), filename+":")
	})
	expected := []string{
		`11:3: verb badYear: invalid cron expression "* * * * * 9999": failed to parse cron expression syntax error in year field: '9999'`,
		`16:3: verb allZeroes: invalid cron expression "0 0 0 0 0": failed to parse cron expression syntax error in day-of-month field: '0'`,
	}
	assert.Equal(t, expected, actual)
}
