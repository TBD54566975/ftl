package modulecontext

import (
	"context"
	"testing"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	cf "github.com/TBD54566975/ftl/common/configuration"
)

func TestReadLatestValue(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	moduleName := "test"
	b := NewBuilder(moduleName)
	b = b.AddConfig("c", "c-value1")
	b = b.AddConfig("c", "c-value2")
	b = b.AddSecret("s", "s-value1")
	b = b.AddSecret("s", "s-value2")
	b = b.AddDSN("d", DBTypePostgres, "postgres://localhost:54320/echo1?sslmode=disable&user=postgres&password=secret")
	b = b.AddDSN("d", DBTypePostgres, "postgres://localhost:54320/echo2?sslmode=disable&user=postgres&password=secret")
	moduleCtx, err := b.Build(ctx)
	assert.NoError(t, err, "there should be no build errors")
	ctx = moduleCtx.ApplyToContext(ctx)

	cm := cf.ConfigFromContext(ctx)
	sm := cf.SecretsFromContext(ctx)

	var str string

	err = cm.Get(ctx, cf.Ref{Module: optional.Some(moduleName), Name: "c"}, &str)
	assert.NoError(t, err, "could not read config value")
	assert.Equal(t, "c-value2", str, "latest config value should be read")

	err = sm.Get(ctx, cf.Ref{Module: optional.Some(moduleName), Name: "s"}, &str)
	assert.NoError(t, err, "could not read secret value")
	assert.Equal(t, "s-value2", str, "latest secret value should be read")

	assert.Equal(t, moduleCtx.dbProvider.entries["d"].dsn, "postgres://localhost:54320/echo2?sslmode=disable&user=postgres&password=secret", "latest DSN should be read")
}
