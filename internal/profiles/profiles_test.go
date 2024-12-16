package profiles_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"

	"github.com/block/ftl"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/providers"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/profiles"
)

func TestProfile(t *testing.T) {
	root := t.TempDir()

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	projectConfig := profiles.ProjectConfig{
		Root:          root,
		Realm:         "test",
		FTLMinVersion: ftl.Version,
		ModuleRoots:   []string{"."},
		NoGit:         true,
	}
	sr := providers.NewRegistry[configuration.Secrets]()
	sr.Register(providers.NewInlineFactory[configuration.Secrets]())
	cr := providers.NewRegistry[configuration.Configuration]()
	cr.Register(providers.NewInlineFactory[configuration.Configuration]())

	_, err := profiles.Init(projectConfig, sr, cr)
	assert.NoError(t, err)

	project, err := profiles.Open(root, sr, cr)
	assert.NoError(t, err)

	profile, err := project.Load(ctx, "local")
	assert.NoError(t, err)

	assert.Equal(t, profiles.Config{
		Name:     "local",
		Endpoint: must.Get(url.Parse("http://localhost:8892")),
	}, profile.Config())

	assert.Equal(t, profiles.ProjectConfig{
		Root:           root,
		Realm:          "test",
		FTLMinVersion:  ftl.Version,
		ModuleRoots:    []string{"."},
		NoGit:          true,
		DefaultProfile: "local",
	}, profile.ProjectConfig())

	cm := profile.ConfigurationManager()
	passwordKey := configuration.NewRef("echo", "password")
	err = cm.Set(ctx, passwordKey, "hello")
	assert.NoError(t, err)

	var passwordValue string
	err = cm.Get(ctx, passwordKey, &passwordValue)
	assert.NoError(t, err)

	assert.Equal(t, "hello", passwordValue)
}
