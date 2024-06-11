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
		Commands: Commands{
			Startup: []string{"echo 'Executing global pre-build command'"},
		},
		filePaths: []string{"testdata/ftl-project.toml"},
	}

	assert.Equal(t, expected, actual)
}

func TestProjectLoadConfig(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		err   string
	}{
		{name: "SingleValid", paths: []string{"testdata/ftl-project.toml"}},
		{name: "MultipleValid", paths: []string{"testdata/ftl-project.toml", "testdata/go/configs-ftl-project.toml"}},
		{name: "IsNonExistent", paths: []string{"testdata/ftl-project-nonexistent.toml"}, err: "no such file or directory"},
		{name: "ContainsNonExistent", paths: []string{"testdata/ftl-project.toml", "testdata/ftl-project-nonexistent.toml"}, err: "no such file or directory"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := LoadConfig(log.ContextWithNewDefaultLogger(context.Background()), test.paths)
			if test.err != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.err)
			} else {
				assert.NoError(t, err)

				// Check that all test.paths exist in config.FilePaths
				assert.Equal(t, len(test.paths), len(config.FilePaths()))
				for _, path := range test.paths {
					found := false
					for _, configPath := range config.FilePaths() {
						if path == configPath {
							found = true
							break
						}
					}
					assert.True(t, found, "expected path %q not found in config.FilePaths", path)
				}
			}
		})
	}
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
