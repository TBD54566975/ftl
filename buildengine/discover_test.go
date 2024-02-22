package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/common/moduleconfig"
)

func TestDiscoverModules(t *testing.T) {
	modules, err := DiscoverModules("testdata/modules")
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
