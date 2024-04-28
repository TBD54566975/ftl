package modulecontext

import (
	"context"
	"fmt"
	"os"
	"os/exec" //nolint:depguard
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestConfigPriority(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	moduleName := "test"

	cp := cf.InlineProvider[cf.Configuration]{}
	cr := cf.NewInMemoryResolver[cf.Configuration]()
	cm, err := cf.New(ctx, cr, []cf.Provider[cf.Configuration]{cp})
	assert.NoError(t, err)
	ctx = cf.ContextWithConfig(ctx, cm)

	sp := cf.InlineProvider[cf.Secrets]{}
	sr := cf.NewInMemoryResolver[cf.Secrets]()
	sm, err := cf.New(ctx, sr, []cf.Provider[cf.Secrets]{sp})
	assert.NoError(t, err)
	ctx = cf.ContextWithSecrets(ctx, sm)

	// Set 50 configs and 50 global configs
	// It's hard to tell if module config beats global configs because we are dealing with unordered maps, or because the logic is correct
	// Repeating it 50 times hopefully gives us a good chance of catching inconsistencies
	for i := range 50 {
		key := fmt.Sprintf("key%d", i)

		strValue := "HelloWorld"
		globalStrValue := "GlobalHelloWorld"
		assert.NoError(t, cm.Set(ctx, cf.Ref{Module: optional.Some(moduleName), Name: key}, strValue))
		assert.NoError(t, cm.Set(ctx, cf.Ref{Module: optional.None[string](), Name: key}, globalStrValue))
	}

	moduleContext := New(moduleName, cm, sm, NewDBProvider())

	response, err := moduleContext.ToProto(ctx)
	assert.NoError(t, err)

	for i := range 50 {
		key := fmt.Sprintf("key%d", i)
		assert.Equal(t, `"HelloWorld"`, string(response.Configs[key]), "module configs should beat global configs")
	}
}

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

	moduleContext, err := FromEnvironment(ctx, "echo")
	assert.NoError(t, err)

	response, err := moduleContext.ToProto(ctx)
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
