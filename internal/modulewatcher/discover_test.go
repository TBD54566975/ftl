package modulewatcher

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
			Deploy:    []string{"main", "launch"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
		},
		{
			Dir:       "testdata/another",
			Language:  "go",
			Realm:     "home",
			Module:    "another",
			Deploy:    []string{"main", "launch"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
		},
		{
			Dir:       "testdata/depcycle1",
			Language:  "go",
			Realm:     "home",
			Module:    "depcycle1",
			Deploy:    []string{"main", "launch"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
		},
		{
			Dir:       "testdata/depcycle2",
			Language:  "go",
			Realm:     "home",
			Module:    "depcycle2",
			Deploy:    []string{"main", "launch"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
		},
		{

			Dir:      "testdata/echokotlin",
			Language: "kotlin",
			Realm:    "home",
			Module:   "echo",
			Build:    "mvn -B package",
			Deploy: []string{
				"quarkus-app",
				"launch",
			},
			DeployDir:          "target",
			GeneratedSchemaDir: "src/main/ftl-module-schema",
			Errors:             "errors.pb",
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
				"main",
				"launch",
			},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
		},
		{

			Dir:      "testdata/externalkotlin",
			Language: "kotlin",
			Realm:    "home",
			Module:   "externalkotlin",
			Build:    "mvn -B package",
			Deploy: []string{
				"quarkus-app",
				"launch",
			},
			DeployDir:          "target",
			GeneratedSchemaDir: "src/main/ftl-module-schema",
			Errors:             "errors.pb",
			Java: moduleconfig.ModuleJavaConfig{
				BuildTool: "maven",
			},
		},
		{

			Dir:       "testdata/integer",
			Language:  "go",
			Realm:     "home",
			Module:    "integer",
			Deploy:    []string{"main", "launch"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
		},
		{

			Dir:       "testdata/other",
			Language:  "go",
			Realm:     "home",
			Module:    "other",
			Deploy:    []string{"main", "launch"},
			DeployDir: ".ftl",
			Errors:    "errors.pb",
		},
	}

	assert.Equal(t, expected, modules)
}
