package buildengine

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/pubsub"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
)

type rustPlugin struct {
	*internalPlugin
}

var _ = LanguagePlugin(&rustPlugin{})

func newRustPlugin(ctx context.Context, config moduleconfig.AbsModuleConfig) *rustPlugin {
	internal := newInternalPlugin(ctx, config, buildRust)
	return &rustPlugin{
		internalPlugin: internal,
	}
}

func (p *rustPlugin) Updates() *pubsub.Topic[PluginEvent] {
	return p.internalPlugin.updates
}

func (p *rustPlugin) Kill(ctx context.Context) error {
	p.cancel()
	return nil
}

func (p *rustPlugin) CreateModule(ctx context.Context, config moduleconfig.AbsModuleConfig, includeBinDir bool, replacements map[string]string, group string) error {
	return fmt.Errorf("not implemented")
}

func (p *rustPlugin) GetDependencies(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *rustPlugin) Build(ctx context.Context, projectRoot string, config moduleconfig.AbsModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool) (BuildResult, error) {
	return p.internalPlugin.build(ctx, projectRoot, config, sch, buildEnv, devMode)
}

func buildRust(ctx context.Context, projectRoot string, config moduleconfig.AbsModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool, transaction ModifyFilesTransaction) error {
	logger := log.FromContext(ctx)

	logger.Debugf("Using build command '%s'", config.Build)
	err := exec.Command(ctx, log.Debug, config.Dir+"/_ftl", "bash", "-c", config.Build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", config.Module, err)
	}
	return nil
}
