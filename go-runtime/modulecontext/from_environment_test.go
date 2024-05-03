package modulecontext

import (
	"context"
	"os"
	"os/exec" //nolint:depguard
	"path/filepath"
	"testing"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
)

func TestFromEnvironment(t *testing.T) {
	// Setup a git repo with a ftl-project.toml file with known values.
	dir := t.TempDir()
	cmd := exec.Command("git", "init", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	assert.NoError(t, err)

	data, err := os.ReadFile("testdata/ftl-project.toml")
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "ftl-project.toml"), data, 0600)
	assert.NoError(t, err)

	t.Setenv("FTL_POSTGRES_DSN_ECHO_ECHO", "postgres://echo:echo@localhost:5432/echo")

	// Move into the temp git repo.
	oldwd, err := os.Getwd()
	assert.NoError(t, err)

	assert.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { assert.NoError(t, os.Chdir(oldwd)) })

	ctx := log.ContextWithNewDefaultLogger(context.Background())

	moduleContext, err := New("echo").UpdateFromEnvironment(ctx)
	assert.NoError(t, err)

	response, err := moduleContext.ToProto()
	assert.NoError(t, err)

	assert.Equal(t, &ftlv1.ModuleContextResponse{
		Module:  "echo",
		Configs: map[string][]uint8{"foo": []byte(`"bar"`)},
		Secrets: map[string][]uint8{"foo": []byte(`"bar"`)},
		Databases: []*ftlv1.ModuleContextResponse_DSN{
			{Name: "echo", Dsn: "postgres://echo:echo@localhost:5432/echo"},
		},
	}, response)
}
