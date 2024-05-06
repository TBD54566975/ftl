package modulecontext

import (
	"context" //nolint:depguard
	"testing"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
)

func TestFromEnvironment(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	t.Setenv("FTL_POSTGRES_DSN_ECHO_ECHO", "postgres://echo:echo@localhost:5432/echo")

	databases, err := DatabasesFromEnvironment(ctx, "echo")
	assert.NoError(t, err)

	response := NewBuilder("echo").AddDatabases(databases).Build().ToProto()
	assert.Equal(t, &ftlv1.ModuleContextResponse{
		Module:  "echo",
		Configs: map[string][]byte{},
		Secrets: map[string][]byte{},
		Databases: []*ftlv1.ModuleContextResponse_DSN{
			{Name: "echo", Dsn: "postgres://echo:echo@localhost:5432/echo"},
		},
	}, response)
}
