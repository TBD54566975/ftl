package profiles_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/profiles"
)

func TestProfile(t *testing.T) {
	root := t.TempDir()

	ctx := context.Background()
	project := profiles.ProjectConfig{
		Root:          root,
		Realm:         "test",
		FTLMinVersion: ftl.Version,
		ModuleRoots:   []string{"."},
		NoGit:         true,
	}
	err := profiles.Init(project)
	assert.NoError(t, err)

	sr := providers.NewRegistry[configuration.Secrets]()
	sr.Register(providers.NewInlineFactory[configuration.Secrets]())
	cr := providers.NewRegistry[configuration.Configuration]()
	cr.Register(providers.NewInlineFactory[configuration.Configuration]())

	local, err := profiles.Load(ctx, sr, cr, root, "local")
	assert.NoError(t, err)

	assert.Equal(t, profiles.Config{
		Name:     "local",
		Endpoint: must.Get(url.Parse("http://localhost:8892")),
	}, local.Config())

	assert.Equal(t, profiles.ProjectConfig{
		Root:           root,
		Realm:          "test",
		FTLMinVersion:  ftl.Version,
		ModuleRoots:    []string{"."},
		NoGit:          true,
		DefaultProfile: "local",
	}, local.ProjectConfig())

	cm := local.ConfigurationManager()
	passwordKey := configuration.NewRef("echo", "password")
	err = cm.Set(ctx, providers.InlineProviderKey, passwordKey, "hello")
	assert.NoError(t, err)

	var passwordValue string
	err = cm.Get(ctx, passwordKey, &passwordValue)
	assert.NoError(t, err)

	assert.Equal(t, "hello", passwordValue)
}
