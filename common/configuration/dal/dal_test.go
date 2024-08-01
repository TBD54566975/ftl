package dal

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/log"
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

func TestModuleSecrets(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)
	assert.NotZero(t, dal)

	err = dal.SetModuleSecretURL(ctx, optional.Some("echo"), "my_secret", "asm://echo.my_secret")
	assert.NoError(t, err)
	err = dal.SetModuleSecretURL(ctx, optional.Some("echo"), "my_secret", "asm://echo.my_secret")
	assert.NoError(t, err)
}
