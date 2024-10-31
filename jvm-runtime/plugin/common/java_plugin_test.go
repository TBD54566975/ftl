package common

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/buildengine/languageplugin"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
)

func TestExtractModuleDepsKotlin(t *testing.T) {
	deps, err := extractKotlinFTLImports("test", "testdata/kotlin/alpha")
	assert.NoError(t, err)
	assert.Equal(t, []string{"builtin", "other"}, deps)
}

func TestJavaConfigDefaults(t *testing.T) {
	watch := []string{
		"pom.xml",
		"src/**",
		"build/generated",
		"target/generated-sources",
	}
	for _, tt := range []struct {
		language string
		dir      string
		expected moduleconfig.CustomDefaults
	}{
		{
			language: "kotlin",
			dir:      "testdata/kotlin/echo",
			expected: moduleconfig.CustomDefaults{
				Build:              optional.Some("mvn -B package"),
				DeployDir:          "target",
				GeneratedSchemaDir: optional.Some("src/main/ftl-module-schema"),
				Watch:              watch,
				LanguageConfig: map[string]any{
					"build-tool": "maven",
				},
			},
		},
		{
			language: "kotlin",
			dir:      "testdata/kotlin/external",
			expected: moduleconfig.CustomDefaults{
				Build:              optional.Some("mvn -B package"),
				DeployDir:          "target",
				GeneratedSchemaDir: optional.Some("src/main/ftl-module-schema"),
				Watch:              watch,
				LanguageConfig: map[string]any{
					"build-tool": "maven",
				},
			},
		},
	} {
		t.Run(tt.dir, func(t *testing.T) {

			ctx := context.Background()
			logger := log.Configure(os.Stderr, log.Config{Level: log.Debug})
			ctx = log.ContextWithLogger(ctx, logger)
			dir, err := filepath.Abs(tt.dir)
			assert.NoError(t, err)

			baseBind, err := url.Parse("http://127.0.0.1:8893")
			assert.NoError(t, err)
			allocator, err := bind.NewBindAllocator(baseBind, 0)
			assert.NoError(t, err)
			plugin, err := languageplugin.New(ctx, allocator, "java", "test")
			assert.NoError(t, err)

			defaults, err := plugin.ModuleConfigDefaults(ctx, dir)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, defaults)
		})
	}
}