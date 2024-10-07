package languageplugin

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/alecthomas/assert/v2"
)

func TestParseImportsFromTestData(t *testing.T) {
	testFilePath := filepath.Join("testdata", "imports.go")
	expectedImports := []string{"fmt", "os"}
	imports, err := parseImports(testFilePath)
	if err != nil {
		t.Fatalf("Failed to parse imports: %v", err)
	}

	if !reflect.DeepEqual(imports, expectedImports) {
		t.Errorf("parseImports() got = %v, want %v", imports, expectedImports)
	}
}

func TestExtractModuleDepsGo(t *testing.T) {
	ctx := context.Background()
	dir, err := filepath.Abs("../testdata/alpha")
	assert.NoError(t, err)
	uncheckedConfig, err := moduleconfig.LoadConfig(dir)
	assert.NoError(t, err)

	plugin, err := New(ctx, uncheckedConfig.Language)
	assert.NoError(t, err)

	customDefaults, err := plugin.ModuleConfigDefaults(ctx, uncheckedConfig.Dir)
	assert.NoError(t, err)

	config, err := uncheckedConfig.FillDefaultsAndValidate(customDefaults)
	assert.NoError(t, err)

	deps, err := plugin.GetDependencies(ctx, config)
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "other"}, deps)
}

func TestGoConfigDefaults(t *testing.T) {
	for _, tt := range []struct {
		dir      string
		expected moduleconfig.CustomDefaults
	}{
		{
			dir: "../testdata/alpha",
			expected: moduleconfig.CustomDefaults{
				Deploy: []string{
					"main",
					"launch",
				},
				DeployDir: ".ftl",
				Watch: []string{
					"**/*.go",
					"go.mod",
					"go.sum",
					"../../../../go-runtime/ftl/**/*.go",
				},
			},
		},
		{
			dir: "../testdata/another",
			expected: moduleconfig.CustomDefaults{
				Deploy: []string{
					"main",
					"launch",
				},
				DeployDir: ".ftl",
				Watch: []string{
					"**/*.go",
					"go.mod",
					"go.sum",
					"../../../../go-runtime/ftl/**/*.go",
					"../../../../go-runtime/schema/testdata/**/*.go",
				},
			},
		},
	} {
		t.Run(tt.dir, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			dir, err := filepath.Abs(tt.dir)
			assert.NoError(t, err)

			plugin, err := New(ctx, "go")
			assert.NoError(t, err)

			defaults, err := plugin.ModuleConfigDefaults(ctx, dir)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, defaults)
		})
	}
}
