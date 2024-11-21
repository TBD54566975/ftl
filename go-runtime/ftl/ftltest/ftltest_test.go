package ftltest

import (
	"context"
	"fmt"
	"testing"
	_ "unsafe"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
)

func PanicsWithErr(t testing.TB, substr string, fn func()) {
	t.Helper()
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("Expected panic, but got nil")
		}

		errStr := fmt.Sprintf("%v", err)
		assert.Contains(t, errStr, substr, "Expected panic message to contain %q, but got %q", substr, errStr)
	}()
	fn()
}

func TestFtlTestProjectNotLoadedInContext(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx = contextWithFakeFTL(ctx)

	// This should panic suggesting to use ftltest.WithDefaultProjectFile()
	PanicsWithErr(t, "ftltest.WithDefaultProjectFile()", func() {
		_ = ftl.Secret[string]{Ref: reflection.Ref{Module: "test", Name: "moo"}}.Get(ctx)
	})
	PanicsWithErr(t, "ftltest.WithDefaultProjectFile()", func() {
		_ = ftl.Config[string]{Ref: reflection.Ref{Module: "test", Name: "moo"}}.Get(ctx)
	})
}

func TestFtlTextContextExtension(t *testing.T) {
	withFakeModule(t, "ftl/test")

	t.Run("extends with a new project file", func(t *testing.T) {
		original := Context(WithProjectFile("testdata/go/wrapped/ftl-project.toml"))
		extended := SubContext(original, WithProjectFile("testdata/go/wrapped/ftl-project-test-1.toml"))

		var config string
		assert.NoError(t, internal.FromContext(original).(*fakeFTL).GetConfig(original, "config", &config)) //nolint:forcetypeassert
		assert.Equal(t, "bazbaz", config, "does not change the original context")
		assert.NoError(t, internal.FromContext(extended).(*fakeFTL).GetConfig(extended, "config", &config)) //nolint:forcetypeassert
		assert.Equal(t, "foobar", config, "overwrites configuration values from the new file")
	})
	t.Run("extends with a new config value", func(t *testing.T) {
		configA := ftl.Config[string]{Ref: reflection.Ref{Module: "ftl/test", Name: "configA"}}
		configB := ftl.Config[string]{Ref: reflection.Ref{Module: "ftl/test", Name: "configB"}}

		original := Context(WithConfig(configA, "a"), WithConfig(configB, "b"))
		extended := SubContext(original, WithConfig(configA, "a.2"))

		var config string
		assert.NoError(t, internal.FromContext(original).(*fakeFTL).GetConfig(original, "configA", &config)) //nolint:forcetypeassert
		assert.Equal(t, "a", config, "does not change the original context")
		assert.NoError(t, internal.FromContext(extended).(*fakeFTL).GetConfig(extended, "configA", &config)) //nolint:forcetypeassert
		assert.Equal(t, "a.2", config, "overwrites configuration values from the new file")
		assert.NoError(t, internal.FromContext(extended).(*fakeFTL).GetConfig(extended, "configB", &config)) //nolint:forcetypeassert
		assert.Equal(t, "b", config, "retains other config from the original context")
	})
	t.Run("extends with a new secret value", func(t *testing.T) {
		secretA := ftl.Secret[string]{Ref: reflection.Ref{Module: "ftl/test", Name: "secretA"}}
		secretB := ftl.Secret[string]{Ref: reflection.Ref{Module: "ftl/test", Name: "secretB"}}

		original := Context(WithSecret(secretA, "a"), WithSecret(secretB, "b"))
		extended := SubContext(original, WithSecret(secretA, "a.2"))

		var config string
		assert.NoError(t, internal.FromContext(original).(*fakeFTL).GetSecret(original, "secretA", &config)) //nolint:forcetypeassert
		assert.Equal(t, "a", config, "does not change the original context")
		assert.NoError(t, internal.FromContext(extended).(*fakeFTL).GetSecret(extended, "secretA", &config)) //nolint:forcetypeassert
		assert.Equal(t, "a.2", config, "overwrites secret values from the new file")
		assert.NoError(t, internal.FromContext(extended).(*fakeFTL).GetSecret(extended, "secretB", &config)) //nolint:forcetypeassert
		assert.Equal(t, "b", config, "retains other secret from the original context")
	})
	t.Run("retains the existing context.Context state", func(t *testing.T) {
		type keyType string
		original := context.WithValue(Context(), keyType("key"), "value")
		extended := SubContext(original, WithProjectFile("testdata/go/wrapped/ftl-project.toml"))

		assert.Equal(t, "value", extended.Value(keyType("key")), "keeps context.Context value from the original context")
	})
}

// mock out module reflection to make it testable
func withFakeModule(t *testing.T, name string) {
	t.Helper()
	var previousModuleGetter = moduleGetter
	moduleGetter = func() string {
		return name
	}
	t.Cleanup(func() {
		moduleGetter = previousModuleGetter
	})
}
