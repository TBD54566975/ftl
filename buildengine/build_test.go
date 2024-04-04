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
	expectFail bool,
	assertions []assertion,
) {
	t.Helper()
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{}))
	abs, err := filepath.Abs(bctx.moduleDir)
	assert.NoError(t, err, "Error getting absolute path for module directory")
	module, err := LoadModule(abs)
	assert.NoError(t, err)
	err = Build(ctx, bctx.sch, module)
	if expectFail {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}

	for _, a := range assertions {
		err = a(t, bctx)
		assert.NoError(t, err)
	}

	err = os.RemoveAll(filepath.Join(bctx.moduleDir, bctx.buildDir))
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

func assertBuildProtoErrors(msgs ...string) assertion {
	return func(t testing.TB, bctx buildContext) error {
		t.Helper()
		errorList, err := loadProtoErrors(filepath.Join(bctx.moduleDir, bctx.buildDir))
		assert.NoError(t, err, "Error loading proto errors")

		expected := make([]*schema.Error, 0, len(msgs))
		for _, msg := range msgs {
			expected = append(expected, &schema.Error{Msg: msg})
		}

		// normalize results
		for _, e := range errorList.Errors {
			e.EndColumn = 0
		}

		assert.Equal(t, errorList.Errors, expected, assert.Exclude[schema.Position]())
		return nil
	}
}
