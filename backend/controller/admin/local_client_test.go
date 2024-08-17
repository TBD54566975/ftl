//go:build integration

package admin

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	cf "github.com/TBD54566975/ftl/common/configuration"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestDiskSchemaRetrieverWithBuildArtefact(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("ftl-project-dr.toml"),
		in.WithoutController(),
		in.CopyModule("dischema"),
		in.Build("dischema"),
		func(t testing.TB, ic in.TestContext) {
			dsr := &diskSchemaRetriever{deployRoot: optional.Some[string](ic.WorkingDir())}
			sch, err := dsr.GetActiveSchema(ic.Context)
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
			dsr := &diskSchemaRetriever{}
			_, err := dsr.GetActiveSchema(ic.Context)
			assert.Error(t, err)
		},
	)
}

func TestAdminNoValidationWithNoSchema(t *testing.T) {
	config := tempConfigPath(t, "testdata/ftl-project.toml", "admin")
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cm, err := cf.NewConfigurationManager(ctx, cf.ProjectConfigResolver[cf.Configuration]{Config: config})
	assert.NoError(t, err)

	sm, err := cf.New(ctx,
		cf.ProjectConfigResolver[cf.Secrets]{Config: config},
		[]cf.Provider[cf.Secrets]{
			cf.EnvarProvider[cf.Secrets]{},
			cf.InlineProvider[cf.Secrets]{},
		})
	assert.NoError(t, err)

	dsr := &diskSchemaRetriever{deployRoot: optional.Some(string(t.TempDir()))}
	_, err = dsr.GetActiveSchema(ctx)
	assert.Error(t, err)

	admin := NewAdminService(cm, sm, dsr)
	testSetConfig(t, ctx, admin, "batmobile", "color", "Red", "")
	testSetSecret(t, ctx, admin, "batmobile", "owner", 99, "")
}
