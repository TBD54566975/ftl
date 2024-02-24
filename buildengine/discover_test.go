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
	modules, err := DiscoverModules(ctx, "testdata/modules")
	assert.NoError(t, err)
	expected := []moduleconfig.ModuleConfig{
		{
			Dir:       "testdata/modules/alpha",
			Language:  "go",
			Realm:     "home",
			Module:    "alpha",
			Deploy:    []string{"main"},
			DeployDir: "_ftl",
			Schema:    "schema.pb",
			Watch:     []string{"**/*.go", "go.mod", "go.sum"},
		},
		{
			Dir:       "testdata/modules/another",
			Language:  "go",
			Realm:     "home",
			Module:    "another",
			Deploy:    []string{"main"},
			DeployDir: "_ftl",
			Schema:    "schema.pb",
			Watch:     []string{"**/*.go", "go.mod", "go.sum"},
		},
		{
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
		{
			Dir:       "testdata/modules/other",
			Language:  "go",
			Realm:     "home",
			Module:    "other",
			Deploy:    []string{"main"},
			DeployDir: "_ftl",
			Schema:    "schema.pb",
			Watch:     []string{"**/*.go", "go.mod", "go.sum"},
		},
	}
	assert.Equal(t, expected, modules)
}
