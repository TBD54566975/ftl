package buildengine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/log"
)

type buildContext struct {
	moduleDir string
	buildDir  string
	sch       *schema.Schema
}

type assertion func(t testing.TB, bctx buildContext) error

type mockModifyFilesTransaction struct{}

func (t *mockModifyFilesTransaction) Begin() error {
	return nil
}

func (t *mockModifyFilesTransaction) ModifiedFiles(paths ...string) error {
	return nil
}

func (t *mockModifyFilesTransaction) End() error {
	return nil
}

func testBuild(
	t *testing.T,
	bctx buildContext,
	expectedBuildErrMsg string, // emptystr if no error expected
	assertions []assertion,
) {
	t.Helper()
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{}))
	abs, err := filepath.Abs(bctx.moduleDir)
	assert.NoError(t, err, "Error getting absolute path for module directory")
	module, err := LoadModule(abs)
	assert.NoError(t, err)
	err = Build(ctx, bctx.sch, module, &mockModifyFilesTransaction{})
	if len(expectedBuildErrMsg) > 0 {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedBuildErrMsg)
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

func testBuildClearsBuildDir(t *testing.T, bctx buildContext) {
	t.Helper()
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{}))
	abs, err := filepath.Abs(bctx.moduleDir)
	assert.NoError(t, err, "Error getting absolute path for module directory")

	// build to generate the build directory
	module, err := LoadModule(abs)
	assert.NoError(t, err)
	err = Build(ctx, bctx.sch, module, &mockModifyFilesTransaction{})
	assert.NoError(t, err)

	// create a temporary file in the build directory
	buildDir := filepath.Join(bctx.moduleDir, bctx.buildDir)
	tempFile, err := os.Create(filepath.Join(buildDir, "test-clear-build.tmp"))
	assert.NoError(t, err, "Error creating temporary file in module directory")
	tempFile.Close()

	// build to clear the old build directory
	module, err = LoadModule(abs)
	assert.NoError(t, err)
	err = Build(ctx, bctx.sch, module, &mockModifyFilesTransaction{})
	assert.NoError(t, err)

	// ensure the temporary file was removed
	_, err = os.Stat(filepath.Join(buildDir, "test-clear-build.tmp"))
	assert.Error(t, err, "Build directory was not removed")
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

func assertGeneratedMain(expectedContent string) assertion {
	return func(t testing.TB, bctx buildContext) error {
		t.Helper()
		output := filepath.Join(bctx.moduleDir, bctx.buildDir, "go/main/main.go")
		fileContent, err := os.ReadFile(output)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, string(fileContent))
		return nil
	}
}

func assertBuildProtoErrors(msgs ...string) assertion {
	return func(t testing.TB, bctx buildContext) error {
		t.Helper()
		config, err := moduleconfig.LoadModuleConfig(bctx.moduleDir)
		assert.NoError(t, err, "Error loading module config")
		errorList, err := loadProtoErrors(config)
		assert.NoError(t, err, "Error loading proto errors")

		expected := make([]*schema.Error, 0, len(msgs))
		for _, msg := range msgs {
			expected = append(expected, &schema.Error{Msg: msg, Level: schema.ERROR})
		}

		// normalize results
		for _, e := range errorList.Errors {
			e.EndColumn = 0
		}

		assert.Equal(t, errorList.Errors, expected, assert.Exclude[schema.Position]())
		return nil
	}
}
