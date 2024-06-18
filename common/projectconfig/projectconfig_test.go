package projectconfig

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestProjectConfig(t *testing.T) {
	actual, err := Load(context.Background(), "testdata/ftl-project.toml")
	assert.NoError(t, err)
	expected := Config{
		Path: actual.Path,
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
		ModuleDirs: []string{"a/b/c", "d"},
		Commands: Commands{
			Startup: []string{"echo 'Executing global pre-build command'"},
		},
	}

	assert.Equal(t, expected, actual)
}

func TestProjectLoadConfig(t *testing.T) {
	tests := []struct {
		name  string
		paths string
		err   string
	}{
		{name: "AllValid", paths: "testdata/ftl-project.toml"},
		{name: "IsNonExistent", paths: "testdata/ftl-project-nonexistent.toml", err: "no such file or directory"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(log.ContextWithNewDefaultLogger(context.Background()), test.paths)
			if test.err != "" {
				assert.Contains(t, err.Error(), test.err)
			} else {
				assert.NoError(t, err)
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

	oldVersion := ftl.Version
	t.Cleanup(func() { ftl.Version = oldVersion })

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ftl.Version = test.v
			_, err := Load(log.ContextWithNewDefaultLogger(context.Background()), test.path)
			if !test.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
