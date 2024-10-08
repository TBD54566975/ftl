package languageplugin

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"

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
		Build:     "cargo build",
		DeployDir: "_ftl/target/debug",
		Deploy:    []string{"main"},
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

func buildRust(ctx context.Context, projectRoot string, config moduleconfig.AbsModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool, transaction watch.ModifyFilesTransaction) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Using build command '%s'", config.Build)
	err := exec.Command(ctx, log.Debug, config.Dir+"/_ftl", "bash", "-c", config.Build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", config.Module, err)
	}
	return nil
}
