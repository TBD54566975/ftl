package buildengine

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestDiscoverModules(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	projects, err := DiscoverProjects(ctx, []string{"testdata/projects"}, []string{"testdata/projects/lib", "testdata/projects/libkotlin"})
	assert.NoError(t, err)
	expected := []Project{
		Module{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:       "testdata/projects/alpha",
				Language:  "go",
				Realm:     "home",
				Module:    "alpha",
				Deploy:    []string{"main"},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum", "../../../../go-runtime/ftl/**/*.go"},
			},
		},
		Module{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:       "testdata/projects/another",
				Language:  "go",
				Realm:     "home",
				Module:    "another",
				Deploy:    []string{"main"},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum", "../../../../go-runtime/ftl/**/*.go"},
			},
		},
		Module{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:       "testdata/projects/depcycle1",
				Language:  "go",
				Realm:     "home",
				Module:    "depcycle1",
				Deploy:    []string{"main"},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum"},
			},
		},
		Module{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:       "testdata/projects/depcycle2",
				Language:  "go",
				Realm:     "home",
				Module:    "depcycle2",
				Deploy:    []string{"main"},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum"},
			},
		},
		Module{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:      "testdata/projects/echokotlin",
				Language: "kotlin",
				Realm:    "home",
				Module:   "echo",
				Build:    "mvn -B compile",
				Deploy: []string{
					"main",
					"classes",
					"dependency",
					"classpath.txt",
				},
				DeployDir: "target",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch: []string{
					"pom.xml",
					"src/**",
					"target/generated-sources",
				},
			},
		},
		Module{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:      "testdata/projects/external",
				Language: "go",
				Realm:    "home",
				Module:   "external",
				Build:    "",
				Deploy: []string{
					"main",
				},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch: []string{
					"**/*.go",
					"go.mod",
					"go.sum",
				},
			},
		},
		Module{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:      "testdata/projects/externalkotlin",
				Language: "kotlin",
				Realm:    "home",
				Module:   "externalkotlin",
				Build:    "mvn -B compile",
				Deploy: []string{
					"main",
					"classes",
					"dependency",
					"classpath.txt",
				},
				DeployDir: "target",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch: []string{
					"pom.xml",
					"src/**",
					"target/generated-sources",
				},
			},
		},
		Module{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:       "testdata/projects/highgoversion",
				Language:  "go",
				Realm:     "home",
				Module:    "highgoversion",
				Deploy:    []string{"main"},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum", "../../../../go-runtime/ftl/**/*.go"},
			},
		},
		Module{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:       "testdata/projects/other",
				Language:  "go",
				Realm:     "home",
				Module:    "other",
				Deploy:    []string{"main"},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Errors:    "errors.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum", "../../../../go-runtime/ftl/**/*.go"},
			},
		},
		ExternalLibrary{
			Dir:      "testdata/projects/lib",
			Language: "go",
		},
		ExternalLibrary{
			Dir:      "testdata/projects/libkotlin",
			Language: "kotlin",
		},
	}
	assert.Equal(t, expected, projects)
}
