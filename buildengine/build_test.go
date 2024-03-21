package buildengine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
)

type buildContext struct {
	moduleDir string
	buildDir  string
	sch       *schema.Schema
}

type assertion func(t testing.TB, bctx buildContext) error

func testBuild(
	t *testing.T,
	bctx buildContext,
	assertions []assertion,
) {
	t.Helper()
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{}))
	module, err := LoadModule(bctx.moduleDir)
	assert.NoError(t, err)

	err = Build(ctx, bctx.sch, module)
	assert.NoError(t, err)

	for _, a := range assertions {
		err = a(t, bctx)
		assert.NoError(t, err)
	}

	err = os.RemoveAll(bctx.buildDir)
	assert.NoError(t, err, "Error removing build directory")
}

func assertGeneratedModule(generatedModulePath string, expectedContent string) assertion {
	return func(t testing.TB, bctx buildContext) error {
		t.Helper()
		target := filepath.Join(bctx.moduleDir, bctx.buildDir)
		output := filepath.Join(target, generatedModulePath)

		fileContent, err := os.ReadFile(output)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, string(fileContent))
		return nil
	}
}
