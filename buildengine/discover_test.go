package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/common/moduleconfig"
)

func TestDiscoverModules(t *testing.T) {
	modules, err := discoverModules("testdata/modules")
	assert.NoError(t, err)
	expected := []Module{
		{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:       "testdata/modules/alpha",
				Language:  "go",
				Realm:     "home",
				Module:    "alpha",
				Deploy:    []string{"main"},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum"},
			},
		},
		{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:       "testdata/modules/another",
				Language:  "go",
				Realm:     "home",
				Module:    "another",
				Deploy:    []string{"main"},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum"},
			},
		},
		{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:      "testdata/modules/echokotlin",
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
				Watch: []string{
					"pom.xml",
					"src/**",
					"target/generated-sources",
				},
			},
		},
		{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:      "testdata/modules/external",
				Language: "go",
				Realm:    "home",
				Module:   "external",
				Build:    "",
				Deploy: []string{
					"main",
				},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Watch: []string{
					"**/*.go",
					"go.mod",
					"go.sum",
				},
			},
		},
		{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:       "testdata/modules/other",
				Language:  "go",
				Realm:     "home",
				Module:    "other",
				Deploy:    []string{"main"},
				DeployDir: "_ftl",
				Schema:    "schema.pb",
				Watch:     []string{"**/*.go", "go.mod", "go.sum"},
			},
		},
		{
			ModuleConfig: moduleconfig.ModuleConfig{
				Dir:      "testdata/modules/externalkotlin",
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
				Watch: []string{
					"pom.xml",
					"src/**",
					"target/generated-sources",
				},
			},
		},
	}
	assert.Equal(t, expected, modules)
}
