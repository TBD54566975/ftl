package languageplugin

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/watch"
)

type rustPlugin struct {
	*internalPlugin
}

var _ = LanguagePlugin(&rustPlugin{})

func newRustPlugin(ctx context.Context) *rustPlugin {
	internal := newInternalPlugin(ctx, "rust", buildRust)
	return &rustPlugin{
		internalPlugin: internal,
	}
}

func (p *rustPlugin) ModuleConfigDefaults(ctx context.Context, dir string) (moduleconfig.CustomDefaults, error) {
	return moduleconfig.CustomDefaults{
		Build:     optional.Some("cargo build"),
		DeployDir: "_ftl/target/debug",
		Watch:     []string{"**/*.rs", "Cargo.toml", "Cargo.lock"},
	}, nil
}

func (p *rustPlugin) GetCreateModuleFlags(ctx context.Context) ([]*kong.Flag, error) {
	return []*kong.Flag{}, nil
}

func (p *rustPlugin) CreateModule(ctx context.Context, projConfig projectconfig.Config, c moduleconfig.ModuleConfig, flags map[string]string) error {
	return fmt.Errorf("not implemented")
}

func (p *rustPlugin) GetDependencies(ctx context.Context, config moduleconfig.ModuleConfig) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func buildRust(ctx context.Context, projectRoot string, bctx BuildContext, buildEnv []string, devMode bool, transaction watch.ModifyFilesTransaction) (BuildResult, error) {
	config := bctx.Config.Abs()
	logger := log.FromContext(ctx)
	logger.Debugf("Using build command '%s'", config.Build)
	err := exec.Command(ctx, log.Debug, config.Dir+"/_ftl", "bash", "-c", config.Build).RunBuffered(ctx)
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to build module %q: %w", config.Module, err)
	}
	buildErrs, err := loadProtoErrors(config)
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to load build errors: %w", err)
	}
	result := BuildResult{
		Errors: buildErrs,
	}
	if builderrors.ContainsTerminalError(buildErrs) {
		// skip reading schema
		return result, nil
	}

	moduleSchema, err := schema.ModuleFromProtoFile(config.Schema())
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to read schema for module: %w", err)
	}

	result.Schema = moduleSchema
	result.Deploy = []string{"main"}

	return result, nil
}
