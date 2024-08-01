package dal

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	libdal "github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestDALConfiguration(t *testing.T) {
	t.Run("ModuleConfiguration", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleConfiguration(ctx, optional.Some("echo"), "my_config", []byte(`""`))
		assert.NoError(t, err)

		value, err := dal.GetModuleConfiguration(ctx, optional.Some("echo"), "my_config")
		assert.NoError(t, err)
		assert.Equal(t, []byte(`""`), value)

		err = dal.UnsetModuleConfiguration(ctx, optional.Some("echo"), "my_config")
		assert.NoError(t, err)

		value, err = dal.GetModuleConfiguration(ctx, optional.Some("echo"), "my_config")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, libdal.ErrNotFound))
		assert.Zero(t, value)
	})

	t.Run("GlobalConfiguration", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleConfiguration(ctx, optional.None[string](), "my_config", []byte(`""`))
		assert.NoError(t, err)

		value, err := dal.GetModuleConfiguration(ctx, optional.None[string](), "my_config")
		assert.NoError(t, err)
		assert.Equal(t, []byte(`""`), value)

		err = dal.UnsetModuleConfiguration(ctx, optional.None[string](), "my_config")
		assert.NoError(t, err)

		value, err = dal.GetModuleConfiguration(ctx, optional.None[string](), "my_config")
		fmt.Printf("value: %v\n", value)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, libdal.ErrNotFound))
		assert.Zero(t, value)
	})

	t.Run("SetSameGlobalConfigTwice", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleConfiguration(ctx, optional.None[string](), "my_config", []byte(`""`))
		assert.NoError(t, err)

		err = dal.SetModuleConfiguration(ctx, optional.None[string](), "my_config", []byte(`"hehe"`))
		assert.NoError(t, err)

		value, err := dal.GetModuleConfiguration(ctx, optional.None[string](), "my_config")
		assert.NoError(t, err)
		assert.Equal(t, []byte(`"hehe"`), value)
	})

	t.Run("SetModuleOverridesGlobal", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleConfiguration(ctx, optional.None[string](), "my_config", []byte(`""`))
		assert.NoError(t, err)
		err = dal.SetModuleConfiguration(ctx, optional.Some("echo"), "my_config", []byte(`"hehe"`))
		assert.NoError(t, err)

		value, err := dal.GetModuleConfiguration(ctx, optional.Some("echo"), "my_config")
		assert.NoError(t, err)
		assert.Equal(t, []byte(`"hehe"`), value)
	})

	t.Run("HandlesConflicts", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleConfiguration(ctx, optional.Some("echo"), "my_config", []byte(`""`))
		assert.NoError(t, err)
		err = dal.SetModuleConfiguration(ctx, optional.Some("echo"), "my_config", []byte(`""`))
		assert.NoError(t, err)

		err = dal.SetModuleConfiguration(ctx, optional.None[string](), "my_config", []byte(`""`))
		assert.NoError(t, err)
		err = dal.SetModuleConfiguration(ctx, optional.None[string](), "my_config", []byte(`"hehe"`))
		assert.NoError(t, err)

		value, err := dal.GetModuleConfiguration(ctx, optional.None[string](), "my_config")
		assert.NoError(t, err)
		assert.Equal(t, []byte(`"hehe"`), value)
	})
}

func TestDALSecrets(t *testing.T) {
	t.Run("ModuleSecret", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleSecretURL(ctx, optional.Some("echo"), "my_secret", "http://example.com")
		assert.NoError(t, err)

		value, err := dal.GetModuleSecretURL(ctx, optional.Some("echo"), "my_secret")
		assert.NoError(t, err)
		assert.Equal(t, "http://example.com", value)

		err = dal.UnsetModuleSecret(ctx, optional.Some("echo"), "my_secret")
		assert.NoError(t, err)

		value, err = dal.GetModuleSecretURL(ctx, optional.Some("echo"), "my_secret")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, libdal.ErrNotFound))
		assert.Zero(t, value)
	})

	t.Run("GlobalSecret", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleSecretURL(ctx, optional.None[string](), "my_secret", "http://example.com")
		assert.NoError(t, err)

		value, err := dal.GetModuleSecretURL(ctx, optional.None[string](), "my_secret")
		assert.NoError(t, err)
		assert.Equal(t, "http://example.com", value)

		err = dal.UnsetModuleSecret(ctx, optional.None[string](), "my_secret")
		assert.NoError(t, err)

		value, err = dal.GetModuleSecretURL(ctx, optional.None[string](), "my_secret")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, libdal.ErrNotFound))
		assert.Zero(t, value)
	})

	t.Run("SetSameGlobalSecretTwice", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleSecretURL(ctx, optional.None[string](), "my_secret", "http://example.com")
		assert.NoError(t, err)

		err = dal.SetModuleSecretURL(ctx, optional.None[string](), "my_secret", "http://example2.com")
		assert.NoError(t, err)

		value, err := dal.GetModuleSecretURL(ctx, optional.None[string](), "my_secret")
		assert.NoError(t, err)
		assert.Equal(t, "http://example2.com", value)
	})

	t.Run("SetModuleOverridesGlobal", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleSecretURL(ctx, optional.None[string](), "my_secret", "http://example.com")
		assert.NoError(t, err)
		err = dal.SetModuleSecretURL(ctx, optional.Some("echo"), "my_secret", "http://example2.com")
		assert.NoError(t, err)

		value, err := dal.GetModuleSecretURL(ctx, optional.Some("echo"), "my_secret")
		assert.NoError(t, err)
		assert.Equal(t, "http://example2.com", value)
	})

	t.Run("HandlesConflicts", func(t *testing.T) {
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		conn := sqltest.OpenForTesting(ctx, t)
		dal, err := New(ctx, conn)
		assert.NoError(t, err)
		assert.NotZero(t, dal)

		err = dal.SetModuleSecretURL(ctx, optional.Some("echo"), "my_secret", "http://example.com")
		assert.NoError(t, err)
		err = dal.SetModuleSecretURL(ctx, optional.Some("echo"), "my_secret", "http://example2.com")
		assert.NoError(t, err)

		value, err := dal.GetModuleSecretURL(ctx, optional.Some("echo"), "my_secret")
		assert.NoError(t, err)
		assert.Equal(t, "http://example2.com", value)

		err = dal.SetModuleSecretURL(ctx, optional.None[string](), "my_secret", "http://example.com")
		assert.NoError(t, err)
		err = dal.SetModuleSecretURL(ctx, optional.None[string](), "my_secret", "http://example2.com")
		assert.NoError(t, err)

		value, err = dal.GetModuleSecretURL(ctx, optional.None[string](), "my_secret")
		assert.NoError(t, err)
		assert.Equal(t, "http://example2.com", value)
	})

}
