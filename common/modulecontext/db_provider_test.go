package modulecontext

import (
	"context"
	"testing"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
)

func TestValidDSN(t *testing.T) {
	dbProvider := NewDBProvider()
	dsn := "postgres://localhost:54320/echo?sslmode=disable&user=postgres&password=secret"
	err := dbProvider.AddDSN("test", DBTypePostgres, dsn)
	assert.NoError(t, err, "expected no error for valid DSN")
	assert.Equal(t, dbProvider.entries["test"].dsn, dsn, "expected DSN to be set and unmodified")
}

func TestGettingAndSettingFromContext(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	dbProvider := NewDBProvider()
	ctx = ContextWithDBProvider(ctx, dbProvider)
	assert.Equal(t, dbProvider, DBProviderFromContext(ctx), "expected dbProvider to be set and retrieved correctly")
}
