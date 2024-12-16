package watch

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

func TestDiscoverModules(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	modules, err := DiscoverModules(ctx, []string{"testdata"})
	assert.NoError(t, err)
	expected := []moduleconfig.UnvalidatedModuleConfig{
		{
			Dir:      "testdata/alpha",
			Language: "go",
			Module:   "alpha",
		},
		{
			Dir:      "testdata/another",
			Language: "go",
			Module:   "another",
		},
		{
			Dir:      "testdata/external",
			Language: "go",
			Module:   "external",
		},
		{
			Dir:      "testdata/integer",
			Language: "go",
			Module:   "integer",
		},
		{
			Dir:      "testdata/other",
			Language: "go",
			Module:   "other",
		},
	}

	assert.Equal(t, expected, modules)
}
