package languageplugin

import (
	"context"
	"fmt"
	"net/url"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/jpillora/backoff"

	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	langconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language/languagepbconnect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
)

const launchTimeout = 10 * time.Second

type State int

const (
	NotStarted State = iota
	Starting
	Started
	Dead
	// TODO: another state for busy?
)

type LanguagePlugin struct {
	bind *url.URL
	path string

	config moduleconfig.ModuleConfig
	state  State

	cmd    *exec.Cmd
	client langconnect.LanguageServiceClient

	// TODO: add a way to remember the state of the last schema we sent to the plugin
}

// TODO: pass in address incrementor so we can relaunch as needed? Or have a way for this plugin to declare that it has errored...
func New(ctx context.Context, path string, bind *url.URL) (*LanguagePlugin, error) {
	config, err := moduleconfig.LoadModuleConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load module config: %w", err)
	}
	return NewWithConfig(ctx, path, bind, config)
}

func NewWithConfig(ctx context.Context, path string, bind *url.URL, config moduleconfig.ModuleConfig) (*LanguagePlugin, error) {
	plugin := &LanguagePlugin{
		bind: bind,
		path: path,

		config: config,

		state:  NotStarted,
		client: rpc.Dial(langconnect.NewLanguageServiceClient, bind.String(), log.Error),
	}

	err := plugin.start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	return plugin, nil
}

// Start launches the plugin and blocks until the plugin is ready.
func (p *LanguagePlugin) start(ctx context.Context) error {
	if p.state != NotStarted {
		return fmt.Errorf("plugin can not start as it has already started")
	}
	p.state = Starting

	logger := log.FromContext(ctx).Scope(p.config.Module)
	// TODO: think more about whether this is a good log level
	// TODO: think more about whether cmd's path should be the current directory, or the module's
	p.cmd = exec.Command(ctx, log.Debug, ".", "ftl-language-"+p.config.Language, "--bind", p.bind.String(), "--path", p.path)

	runCtx, cancel := context.WithCancel(ctx)
	// run the plugin and wait for it to finish executing
	go func() {
		err := p.cmd.RunBuffered(runCtx)
		if err != nil {
			logger.Errorf(err, "language plugin failed")
			cancel()
			p.state = Dead
			// TODO: handle error
		}
	}()
	// kill the plugin when the context is cancelled
	go func() {
		<-runCtx.Done()
		p.Kill()
		p.state = Dead
	}()

	// wait for the plugin to be ready
	if err := p.ping(runCtx); err != nil {
		cancel()
		return fmt.Errorf("failed to ping plugin")
	}

	p.state = Started
	return nil
}

// TODO: make sure we call this
func (p *LanguagePlugin) Kill() error {
	if p.cmd == nil {
		return nil
	}
	return p.cmd.Kill(syscall.SIGABRT)
}

func (p *LanguagePlugin) ping(ctx context.Context) error {
	// TODO: check backoff, too slow?
	retry := backoff.Backoff{}
	heartbeatCtx, cancel := context.WithTimeout(ctx, launchTimeout)
	defer cancel()
	err := rpc.Wait(heartbeatCtx, retry, p.client)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to connect to runner: %w", err))
	}
	return nil
}

// CreateModule creates a new module in the given directory with the given name and language.
func (p *LanguagePlugin) CreateModule(ctx context.Context) error {
	// TODO: check language is supported
	_, err := p.client.CreateModule(ctx, connect.NewRequest(&langpb.CreateModuleRequest{
		Name: p.config.Module,
		Path: p.path,
	}))
	return err
}

// TODO: docs
func (p *LanguagePlugin) GetDependencies(ctx context.Context) ([]string, error) {
	resp, err := p.client.GetDependencies(ctx, connect.NewRequest(&langpb.DependenciesRequest{
		Metadata: metadataProtoFromModuleConfig(p.config),
		Path:     p.path,
	}))
	if err != nil {
		return nil, err
	}
	return resp.Msg.Modules, nil
}

type BuildResult struct {
	Errors []*builderrors.Error
	Schema *schema.Module
	// TODO: ...
}

// TODO: docs
// TODO: add a watch version
func (p *LanguagePlugin) Build(ctx context.Context, sch *schema.Schema, projectPath string) (BuildResult, error) {
	stream, err := p.client.Build(ctx, connect.NewRequest(&langpb.BuildRequest{
		Path:        p.path,
		ProjectPath: projectPath,
		Watch:       false,
		Metadata:    metadataProtoFromModuleConfig(p.config),
		Schema:      sch.ToProto().(*schemapb.Schema), //nolint:forcetypeassert
	}))
	if err != nil {
		return BuildResult{}, err
	}
	defer stream.Close()
	for stream.Receive() {
		msg := stream.Msg()
		switch e := msg.Event.(type) {
		case *langpb.BuildEvent_LogMessage:
			// TODO: handle
		case *langpb.BuildEvent_DependencyUpdate:
			// TODO: handle
		case *langpb.BuildEvent_BuildResult:
			moduleSch, err := schema.ModuleFromProto(e.BuildResult.Module)
			if err != nil {
				return BuildResult{}, fmt.Errorf("failed to parse schema: %w", err)
			}
			errs := langpb.ErrorsFromProto(e.BuildResult.Errors)
			builderrors.SortErrorsByPosition(errs)
			return BuildResult{
				Errors: errs,
				Schema: moduleSch,
				// TODO: ...
			}, nil
		}
	}
	return BuildResult{}, fmt.Errorf("build failed")
}

func metadataProtoFromModuleConfig(config moduleconfig.ModuleConfig) *langpb.Metadata {
	return &langpb.Metadata{
		Name: config.Module,
	}
}
