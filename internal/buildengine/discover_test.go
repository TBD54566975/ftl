package buildengine

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
)

func TestDiscoverModules(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	modules, err := DiscoverModules(ctx, []string{"testdata"})
	assert.NoError(t, err)
	expected := []Module{
		{
			Config: moduleconfig.ModuleConfig{
				Dir:       "testdata/alpha",
				Language:  "go",
				Realm:     "home",
				Module:    "alpha",
				Deploy:    []string{"main"},
				DeployDir: ".ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum", "../../../../go-runtime/ftl/**/*.go"},
			},
		},
		{
			Config: moduleconfig.ModuleConfig{
				Dir:       "testdata/another",
				Language:  "go",
				Realm:     "home",
				Module:    "another",
				Deploy:    []string{"main"},
				DeployDir: ".ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum", "../../../../go-runtime/ftl/**/*.go"},
			},
		},
		{
			Config: moduleconfig.ModuleConfig{
				Dir:       "testdata/depcycle1",
				Language:  "go",
				Realm:     "home",
				Module:    "depcycle1",
				Deploy:    []string{"main"},
				DeployDir: ".ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum"},
			},
		},
		{
			Config: moduleconfig.ModuleConfig{
				Dir:       "testdata/depcycle2",
				Language:  "go",
				Realm:     "home",
				Module:    "depcycle2",
				Deploy:    []string{"main"},
				DeployDir: ".ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum"},
			},
		},
		{
			Config: moduleconfig.ModuleConfig{
				Dir:      "testdata/echokotlin",
				Language: "kotlin",
				Realm:    "home",
				Module:   "echo",
				Build:    "mvn -B package",
				Deploy: []string{
					"main",
					"quarkus-app",
				},
				DeployDir:          "target",
				GeneratedSchemaDir: "src/main/ftl-module-schema",
				Schema:             "schema.pb",
				Errors:             "errors.pb",
				Watch: []string{
					"pom.xml",
					"src/**",
					"target/generated-sources",
				},
				Java: moduleconfig.ModuleJavaConfig{
					BuildTool: "maven",
				},
			},
		},
		{
			Config: moduleconfig.ModuleConfig{
				Dir:      "testdata/external",
				Language: "go",
				Realm:    "home",
				Module:   "external",
				Build:    "",
				Deploy: []string{
					"main",
				},
				DeployDir: ".ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch: []string{
					"**/*.go",
					"go.mod",
					"go.sum",
				},
			},
		},
		{
			Config: moduleconfig.ModuleConfig{
				Dir:      "testdata/externalkotlin",
				Language: "kotlin",
				Realm:    "home",
				Module:   "externalkotlin",
				Build:    "mvn -B package",
				Deploy: []string{
					"main",
					"quarkus-app",
				},
				DeployDir:          "target",
				GeneratedSchemaDir: "src/main/ftl-module-schema",
				Schema:             "schema.pb",
				Errors:             "errors.pb",
				Watch: []string{
					"pom.xml",
					"src/**",
					"target/generated-sources",
				},
				Java: moduleconfig.ModuleJavaConfig{
					BuildTool: "maven",
				},
			},
		},
		{
			Config: moduleconfig.ModuleConfig{
				Dir:       "testdata/integer",
				Language:  "go",
				Realm:     "home",
				Module:    "integer",
				Deploy:    []string{"main"},
				DeployDir: ".ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum"},
			},
		},
		{
			Config: moduleconfig.ModuleConfig{
				Dir:       "testdata/other",
				Language:  "go",
				Realm:     "home",
				Module:    "other",
				Deploy:    []string{"main"},
				DeployDir: ".ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum", "../../../../go-runtime/ftl/**/*.go"},
			},
		},
	}

	assert.Equal(t, expected, modules)
}
