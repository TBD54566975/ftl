package projectconfig

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestProjectConfig(t *testing.T) {
	actual, err := loadFile("testdata/ftl-project.toml")
	assert.NoError(t, err)
	expected := Config{
		Modules: map[string]ConfigAndSecrets{
			"module": {
				Config: map[string]*URL{
					"githubAccessToken": MustParseURL("keychain://githubAccessToken"),
				},
				Secrets: map[string]*URL{
					"companyApiKey": MustParseURL("op://devel/yj3jfj2vzsbiwqabprflnl27lm/companyApiKey"),
					"encryptionKey": MustParseURL("inline://notASensitiveSecret"),
				},
			},
		},
		ModuleDirs:   []string{"a/b/c", "d"},
		ExternalDirs: []string{"e/f", "g/h"},
		Executables: Executables{
			FTL: "ftl",
		},
		Commands: Commands{
			Startup: []string{"echo 'Executing global pre-build command'"},
		},
	}

	assert.Equal(t, expected, actual)
}

func TestProjectConfigChecksMinVersion(t *testing.T) {
	tests := []struct {
		v       string
		wantErr bool
	}{
		{"dev", false},
		{"1.0.0", false},
		{"0.0.1", true},
	}

	for _, test := range tests {
		ftl.Version = test.v
		_, err := LoadConfig(log.ContextWithNewDefaultLogger(context.Background()), []string{"testdata/withMinVersion/ftl-project.toml"})
		if !test.wantErr {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
			_, ok := err.(*ftl.VersionNotSupportedError)
			assert.True(t, ok, "Error should of been of type ftl.VersionNotSupportedError, but instead was: %w", err)
		}
	}
}
