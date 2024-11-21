//go:build integration

package admin

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

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
	dsr := &diskSchemaRetriever{}
	return dsr.GetActiveSchema(ctx)
}

func TestDiskSchemaRetrieverWithBuildArtefact(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("ftl-project-dr.toml"),
		in.WithoutController(),
		in.WithoutProvisioner(),
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
		in.WithoutProvisioner(),
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

	cm, err := manager.New(ctx, routers.ProjectConfig[cf.Configuration]{Config: config}, providers.NewInline[cf.Configuration]())
	assert.NoError(t, err)

	sm, err := manager.New(ctx, routers.ProjectConfig[cf.Secrets]{Config: config}, providers.NewInline[cf.Secrets]())
	assert.NoError(t, err)

	dsr := &diskSchemaRetriever{deployRoot: optional.Some(string(t.TempDir()))}
	_, err = dsr.GetActiveSchema(ctx)
	assert.Error(t, err)

	admin := NewAdminService(cm, sm, dsr)
	testSetConfig(t, ctx, admin, "batmobile", "color", "Red", "")
	testSetSecret(t, ctx, admin, "batmobile", "owner", 99, "")
}
