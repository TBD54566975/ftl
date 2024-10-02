package watch

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
	expected := []moduleconfig.ModuleConfig{
		{
			Dir:       "testdata/alpha",
			Language:  "go",
			Realm:     "home",
			Module:    "alpha",
			Deploy:    []string{"launch", "main"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
			Watch: []string{
				"**/*.go",
				"../../../../go-runtime/ftl/**/*.go",
				"go.mod",
				"go.sum",
			},
		},
		{
			Dir:       "testdata/another",
			Language:  "go",
			Realm:     "home",
			Module:    "another",
			Deploy:    []string{"launch", "main"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
			Watch: []string{
				"**/*.go",
				"../../../../go-runtime/ftl/**/*.go",
				"../../../../go-runtime/schema/testdata/**/*.go",
				"go.mod",
				"go.sum",
			},
		},
		{
			Dir:       "testdata/depcycle1",
			Language:  "go",
			Realm:     "home",
			Module:    "depcycle1",
			Deploy:    []string{"launch", "main"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
			Watch: []string{
				"**/*.go",
				"go.mod",
				"go.sum",
			},
		},
		{
			Dir:       "testdata/depcycle2",
			Language:  "go",
			Realm:     "home",
			Module:    "depcycle2",
			Deploy:    []string{"launch", "main"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
			Watch: []string{
				"**/*.go",
				"go.mod",
				"go.sum",
			},
		},
		{

			Dir:      "testdata/echokotlin",
			Language: "kotlin",
			Realm:    "home",
			Module:   "echo",
			Build:    "mvn -B package",
			Deploy: []string{
				"launch",
				"quarkus-app",
			},
			DeployDir:          "target",
			GeneratedSchemaDir: "src/main/ftl-module-schema",
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
		{

			Dir:      "testdata/external",
			Language: "go",
			Realm:    "home",
			Module:   "external",
			Build:    "",
			Deploy: []string{
				"launch",
				"main",
			},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
			Watch: []string{
				"**/*.go",
				"go.mod",
				"go.sum",
			},
		},
		{

			Dir:      "testdata/externalkotlin",
			Language: "kotlin",
			Realm:    "home",
			Module:   "externalkotlin",
			Build:    "mvn -B package",
			Deploy: []string{
				"launch",
				"quarkus-app",
			},
			DeployDir:          "target",
			GeneratedSchemaDir: "src/main/ftl-module-schema",
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
		{

			Dir:       "testdata/integer",
			Language:  "go",
			Realm:     "home",
			Module:    "integer",
			Deploy:    []string{"launch", "main"},
			DeployDir: ".ftl",
			Watch: []string{
				"**/*.go",
				"go.mod",
				"go.sum",
			},
			Errors: "errors.pb",
		},
		{

			Dir:       "testdata/other",
			Language:  "go",
			Realm:     "home",
			Module:    "other",
			Deploy:    []string{"launch", "main"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
			Watch: []string{
				"**/*.go",
				"../../../../go-runtime/ftl/**/*.go",
				"../../../../go-runtime/schema/testdata/**/*.go",
				"go.mod",
				"go.sum",
			},
		},
	}

	assert.Equal(t, expected, modules)
}
