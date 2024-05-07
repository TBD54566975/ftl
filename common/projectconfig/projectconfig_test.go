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
		Global: ConfigAndSecrets{
			Config: map[string]*URL{
				"ftlEndpoint":          MustParseURL("http://ftlEndpoint"),
				"ftlEndpointAlternate": MustParseURL("http://ftlEndpointAlternate"),
			},
		},
		Modules: map[string]ConfigAndSecrets{
			"module": {
				Config: map[string]*URL{
					"githubAccessToken":      MustParseURL("keychain://githubAccessToken"),
					"someServiceAccessToken": MustParseURL("keychain://someServiceAccessToken"),
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
		name    string
		path    string
		v       string
		wantErr bool
	}{
		{"DevWithMinVersion", "testdata/withMinVersion/ftl-project.toml", "dev", false},
		{"AboveMinVersion", "testdata/withMinVersion/ftl-project.toml", "1.0.0", false},
		{"BelowMinVersion", "testdata/withMinVersion/ftl-project.toml", "0.0.1", true},
		{"DevWithoutMinVersion", "testdata/ftl-project.toml", "dev", false},
		{"AboveWithoutMinVersion", "testdata/ftl-project.toml", "1.0.0", false},
		{"BelowWithoutMinVersion", "testdata/ftl-project.toml", "0.0.1", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ftl.Version = test.v
			_, err := LoadConfig(log.ContextWithNewDefaultLogger(context.Background()), []string{test.path})
			if !test.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
