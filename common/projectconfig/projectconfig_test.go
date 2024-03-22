package projectconfig

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestProjectConfig(t *testing.T) {
	actual, err := Load("testdata/ftl-project.toml")
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
	}

	assert.Equal(t, expected, actual)
}
