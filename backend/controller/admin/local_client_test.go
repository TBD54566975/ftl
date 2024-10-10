//go:build integration

package admin

import (
	"context"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/bind"
	cf "github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/configuration/routers"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"
)

func getDiskSchema(t testing.TB, ctx context.Context) (*schema.Schema, error) {
	t.Helper()

	bindURL, err := url.Parse("http://127.0.0.1:8893")
	assert.NoError(t, err)
	bindAllocator, err := bind.NewBindAllocator(bindURL)
	assert.NoError(t, err)
	dsr := &diskSchemaRetriever{}
	return dsr.GetActiveSchema(ctx, optional.Some(bindAllocator))
}

func TestDiskSchemaRetrieverWithBuildArtefact(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("ftl-project-dr.toml"),
		in.WithoutController(),
		in.CopyModule("dischema"),
		in.Build("dischema"),
		func(t testing.TB, ic in.TestContext) {
			sch, err := getDiskSchema(t, ic.Context)
			assert.NoError(t, err)

			module, ok := sch.Module("dischema").Get()
			assert.Equal(t, ok, true)
			assert.Equal(t, "dischema", module.Name)
		},
	)
}

func TestDiskSchemaRetrieverWithNoSchema(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("ftl-project-dr.toml"),
		in.WithoutController(),
		in.CopyModule("dischema"),
		func(t testing.TB, ic in.TestContext) {
			_, err := getDiskSchema(t, ic.Context)
			assert.Error(t, err)
		},
	)
}

func TestAdminNoValidationWithNoSchema(t *testing.T) {
	config := tempConfigPath(t, "testdata/ftl-project.toml", "admin")
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cm, err := manager.NewConfigurationManager(ctx, routers.ProjectConfig[cf.Configuration]{Config: config})
	assert.NoError(t, err)

	sm, err := manager.New(ctx,
		routers.ProjectConfig[cf.Secrets]{Config: config},
		[]cf.Provider[cf.Secrets]{
			providers.Envar[cf.Secrets]{},
			providers.Inline[cf.Secrets]{},
		})
	assert.NoError(t, err)

	dsr := &diskSchemaRetriever{deployRoot: optional.Some(string(t.TempDir()))}
	_, err = dsr.GetActiveSchema(ctx, optional.None[*bind.BindAllocator]())
	assert.Error(t, err)

	admin := NewAdminService(cm, sm, dsr, optional.None[*bind.BindAllocator]())
	testSetConfig(t, ctx, admin, "batmobile", "color", "Red", "")
	testSetSecret(t, ctx, admin, "batmobile", "owner", 99, "")
}
