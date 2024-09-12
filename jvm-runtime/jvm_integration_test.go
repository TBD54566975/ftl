//go:build integration

package ftl_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	in "github.com/TBD54566975/ftl/internal/integration"

	"github.com/alecthomas/repr"
)

func TestLifecycleJVM(t *testing.T) {
	in.Run(t,
		in.WithLanguages("java", "kotlin"),
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		in.IfLanguage("java", in.Exec("ftl", "new", "java", ".", "echo")),
		in.IfLanguage("kotlin", in.Exec("ftl", "new", "kotlin", ".", "echo")),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!", response)
		}),
	)
}

func TestVerbCalls(t *testing.T) {
	in.Run(t,
		in.WithLanguages("java"),
		in.CopyModule("verbs"),
		in.Deploy("verbs"),
		in.Call("verbs", "anyInput", map[string]string{"name": "Jimmy"}, func(t testing.TB, response string) {
			assert.Equal(t, "Jimmy", response)
		}),
		in.Call("verbs", "anyOutput", "Jimmy", func(t testing.TB, response map[string]string) {
			assert.Equal(t, map[string]string{"name": "Jimmy"}, response)
		}),
	)
}

func TestJVMToGoCall(t *testing.T) {

	exampleObject := TestObject{
		IntField:    43,
		FloatField:  .2,
		StringField: "obj",
		BytesField:  []byte{87, 2, 9},
		BoolField:   true,
		TimeField:   time.Now().UTC(),
		ArrayField:  []string{"foo", "bar"},
		MapField:    map[string]string{"gar": "har"},
	}
	exampleOptionalFieldsObject := TestObjectOptionalFields{
		IntField:    ftl.Some[int](43),
		FloatField:  ftl.Some[float64](.2),
		StringField: ftl.Some[string]("obj"),
		BytesField:  ftl.Some[[]byte]([]byte{87, 2, 9}),
		BoolField:   ftl.Some[bool](true),
		TimeField:   ftl.Some[time.Time](time.Now().UTC()),
		ArrayField:  ftl.Some[[]string]([]string{"foo", "bar"}),
		MapField:    ftl.Some[map[string]string](map[string]string{"gar": "har"}),
	}
	parameterizedObject := ParameterizedType[string]{
		Value:  "foo",
		Array:  []string{"foo"},
		Option: ftl.Some[string]("foo"),
		Map:    map[string]string{"foo": "bar"},
	}

	tests := []in.SubTest{}
	tests = append(tests, PairedTest("emptyVerb", func(module string) in.Action {
		return in.Call(module, "emptyVerb", in.Obj{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, map[string]any{}, response, "expecting empty response, got %s", repr.String(response))
		})
	})...)
	tests = append(tests, PairedTest("sinkVerb", func(module string) in.Action {
		return in.Call(module, "sinkVerb", "ignored", func(t testing.TB, response in.Obj) {
			assert.Equal(t, map[string]any{}, response, "expecting empty response, got %s", repr.String(response))
		})
	})...)
	tests = append(tests, PairedTest("sourceVerb", func(module string) in.Action {
		return in.Call(module, "sourceVerb", in.Obj{}, func(t testing.TB, response string) {
			assert.Equal(t, "Source Verb", response, "expecting empty response, got %s", response)
		})
	})...)
	tests = append(tests, PairedTest("errorEmptyVerb", func(module string) in.Action {
		return in.Fail(
			in.Call(module, "errorEmptyVerb", in.Obj{}, func(t testing.TB, response in.Obj) {
				assert.Equal(t, map[string]any{}, response, "expecting empty response, got %s", repr.String(response))
			}), "verb failed")
	})...)
	tests = append(tests, PairedVerbTest("intVerb", 124)...)
	tests = append(tests, PairedVerbTest("floatVerb", 0.123)...)
	tests = append(tests, PairedVerbTest("stringVerb", "Hello World")...)
	tests = append(tests, PairedVerbTest("bytesVerb", []byte{1, 2, 3, 0, 1})...)
	tests = append(tests, PairedVerbTest("boolVerb", true)...)
	tests = append(tests, PairedVerbTest("stringArrayVerb", []string{"Hello World"})...)
	tests = append(tests, PairedVerbTest("stringMapVerb", map[string]string{"Hello": "World"})...)
	tests = append(tests, PairedTest("timeVerb", func(module string) in.Action {
		now := time.Now().UTC()
		return in.Call(module, "timeVerb", now.Format(time.RFC3339Nano), func(t testing.TB, response string) {
			result, err := time.Parse(time.RFC3339Nano, response)
			assert.NoError(t, err, "time is not a valid RFC3339 time: %s", response)
			assert.Equal(t, now, result, "times not equal %s %s", now, result)
		})
	})...)
	tests = append(tests, PairedVerbTest("testObjectVerb", exampleObject)...)
	tests = append(tests, PairedVerbTest("testObjectOptionalFieldsVerb", exampleOptionalFieldsObject)...)
	// TODO turn on when https://github.com/TBD54566975/ftl/issues/2651 is fixed
	//tests = append(tests, PairedVerbTest("objectMapVerb", map[string]TestObject{"hello": exampleObject})...)
	tests = append(tests, PairedVerbTest("objectArrayVerb", []TestObject{exampleObject})...)
	tests = append(tests, PairedVerbTest("parameterizedObjectVerb", parameterizedObject)...)
	tests = append(tests, PairedVerbTest("optionalIntVerb", -3)...)
	tests = append(tests, PairedVerbTest("optionalFloatVerb", -7.6)...)
	tests = append(tests, PairedVerbTest("optionalStringVerb", "foo")...)
	tests = append(tests, PairedVerbTest("optionalBytesVerb", []byte{134, 255, 0})...)
	tests = append(tests, PairedVerbTest("optionalBoolVerb", false)...)
	tests = append(tests, PairedVerbTest("optionalStringArrayVerb", []string{"foo"})...)
	tests = append(tests, PairedVerbTest("optionalStringMapVerb", map[string]string{"Hello": "World"})...)
	tests = append(tests, PairedTest("optionalTimeVerb", func(module string) in.Action {
		now := time.Now().UTC()
		return in.Call(module, "optionalTimeVerb", now.Format(time.RFC3339Nano), func(t testing.TB, response string) {
			result, err := time.Parse(time.RFC3339Nano, response)
			assert.NoError(t, err, "time is not a valid RFC3339 time: %s", response)
			assert.Equal(t, now, result, "times not equal %s %s", now, result)
		})
	})...)

	tests = append(tests, PairedVerbTest("optionalTestObjectVerb", exampleObject)...)
	tests = append(tests, PairedVerbTest("optionalTestObjectOptionalFieldsVerb", exampleOptionalFieldsObject)...)
	tests = append(tests, PairedVerbTest("externalTypeVerb", "did:web:abc123")...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalIntVerb", ftl.None[int]())...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalFloatVerb", ftl.None[float64]())...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalStringVerb", ftl.None[string]())...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalBytesVerb", ftl.None[[]byte]())...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalBoolVerb", ftl.None[bool]())...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalStringArrayVerb", ftl.None[[]string]())...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalStringMapVerb", ftl.None[map[string]string]())...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalTimeVerb", ftl.None[time.Time]())...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalTestObjectVerb", ftl.None[any]())...)
	// tests = append(tests, PairedPrefixVerbTest("nilvalue", "optionalTestObjectOptionalFieldsVerb", ftl.None[any]())...)

	in.Run(t,
		in.CopyModuleWithLanguage("gomodule", "go"),
		in.CopyModuleWithLanguage("javamodule", "java"),
		in.CopyModuleWithLanguage("kotlinmodule", "kotlin"),
		in.Deploy("gomodule"),
		in.Deploy("javamodule"),
		in.Deploy("kotlinmodule"),
		in.SubTests(tests...),
	)
}

func TestGradle(t *testing.T) {
	in.Run(t,
		in.WithLanguages("java"),
		in.CopyModule("gradle"),
		in.Deploy("gradle"),
		in.Call("gradle", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!", response)
		}),
	)
}

func PairedTest(name string, testFunc func(module string) in.Action) []in.SubTest {
	return []in.SubTest{
		{
			Name:   name + "-go",
			Action: testFunc("gomodule"),
		},
		{
			Name:   name + "-java",
			Action: testFunc("javamodule"),
		},
		{
			Name:   name + "-kotlin",
			Action: testFunc("kotlinmodule"),
		},
	}
}

func VerbTest[T any](verb string, value T) func(module string) in.Action {
	return func(module string) in.Action {
		return in.Call(module, verb, value, func(t testing.TB, response T) {
			assert.Equal(t, value, response, "verb call results not equal %s %s", value, response)
		})
	}
}

func PairedVerbTest[T any](verb string, value T) []in.SubTest {
	return PairedTest(verb, VerbTest[T](verb, value))
}

func PairedPrefixVerbTest[T any](prefex string, verb string, value T) []in.SubTest {
	return PairedTest(prefex+"-"+verb, VerbTest[T](verb, value))
}

type TestObject struct {
	IntField    int               `json:"intField"`
	FloatField  float64           `json:"floatField"`
	StringField string            `json:"stringField"`
	BytesField  []byte            `json:"bytesField"`
	BoolField   bool              `json:"boolField"`
	TimeField   time.Time         `json:"timeField"`
	ArrayField  []string          `json:"arrayField"`
	MapField    map[string]string `json:"mapField"`
}

type TestObjectOptionalFields struct {
	IntField    ftl.Option[int]               `json:"intField"`
	FloatField  ftl.Option[float64]           `json:"floatField"`
	StringField ftl.Option[string]            `json:"stringField"`
	BytesField  ftl.Option[[]byte]            `json:"bytesField"`
	BoolField   ftl.Option[bool]              `json:"boolField"`
	TimeField   ftl.Option[time.Time]         `json:"timeField"`
	ArrayField  ftl.Option[[]string]          `json:"arrayField"`
	MapField    ftl.Option[map[string]string] `json:"mapField"`
}

type ParameterizedType[T any] struct {
	Value  T             `json:"value"`
	Array  []T           `json:"array"`
	Option ftl.Option[T] `json:"option"`
	Map    map[string]T  `json:"map"`
}
