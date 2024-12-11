package goplugin

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"connectrpc.com/connect"

	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/TBD54566975/ftl/common/slices"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/moduleconfig"
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
	dir, err := filepath.Abs("testdata/alpha")
	assert.NoError(t, err)
	uncheckedConfig, err := moduleconfig.LoadConfig(dir)
	assert.NoError(t, err)

	service := New()

	customDefaultsResp, err := service.ModuleConfigDefaults(ctx, connect.NewRequest(&langpb.ModuleConfigDefaultsRequest{
		Dir: uncheckedConfig.Dir,
	}))
	assert.NoError(t, err)

	config, err := uncheckedConfig.FillDefaultsAndValidate(defaultsFromProto(customDefaultsResp.Msg))
	assert.NoError(t, err)

	configProto, err := langpb.ModuleConfigToProto(config.Abs())
	assert.NoError(t, err)
	depsResp, err := service.GetDependencies(ctx, connect.NewRequest(&langpb.GetDependenciesRequest{
		ModuleConfig: configProto,
	}))
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "other"}, depsResp.Msg.Modules)
}

func TestGoConfigDefaults(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		dir      string
		expected moduleconfig.CustomDefaults
	}{
		{
			dir: "testdata/alpha",
			expected: moduleconfig.CustomDefaults{
				DeployDir: ".ftl",
				Watch: []string{
					"**/*.go",
					"go.mod",
					"go.sum",
					"../../../../go-runtime/ftl/**/*.go",
				},
				SQLMigrationDir: "db",
			},
		},
		{
			dir: "testdata/another",
			expected: moduleconfig.CustomDefaults{
				DeployDir: ".ftl",
				Watch: []string{
					"**/*.go",
					"go.mod",
					"go.sum",
					"../../../../go-runtime/ftl/**/*.go",
					"../../../../go-runtime/schema/testdata/**/*.go",
				},
				SQLMigrationDir: "db",
			},
		},
	} {
		t.Run(tt.dir, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			dir, err := filepath.Abs(tt.dir)
			assert.NoError(t, err)

			service := New()

			defaultsResp, err := service.ModuleConfigDefaults(ctx, connect.NewRequest(&langpb.ModuleConfigDefaultsRequest{
				Dir: dir,
			}))
			assert.NoError(t, err)

			defaults := defaultsFromProto(defaultsResp.Msg)
			defaults.Watch = slices.Sort(defaults.Watch)
			tt.expected.Watch = slices.Sort(tt.expected.Watch)
			assert.Equal(t, tt.expected, defaults)
		})
	}
}

func defaultsFromProto(proto *langpb.ModuleConfigDefaultsResponse) moduleconfig.CustomDefaults {
	return moduleconfig.CustomDefaults{
		DeployDir:          proto.DeployDir,
		Watch:              proto.Watch,
		Build:              optional.Ptr(proto.Build),
		DevModeBuild:       optional.Ptr(proto.DevModeBuild),
		GeneratedSchemaDir: optional.Ptr(proto.GeneratedSchemaDir),
		LanguageConfig:     proto.LanguageConfig.AsMap(),
		SQLMigrationDir:    proto.SqlMigrationDir,
	}
}
