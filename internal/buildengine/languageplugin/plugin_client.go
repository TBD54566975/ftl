package languageplugin

import (
	"context"
	"fmt"
	"syscall"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/result"

	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	langconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language/languagepbconnect"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

type streamCancelFunc func()

type pluginClient interface {
	getCreateModuleFlags(ctx context.Context, req *connect.Request[langpb.GetCreateModuleFlagsRequest]) (*connect.Response[langpb.GetCreateModuleFlagsResponse], error)
	createModule(ctx context.Context, req *connect.Request[langpb.CreateModuleRequest]) (*connect.Response[langpb.CreateModuleResponse], error)
	moduleConfigDefaults(ctx context.Context, req *connect.Request[langpb.ModuleConfigDefaultsRequest]) (*connect.Response[langpb.ModuleConfigDefaultsResponse], error)
	getDependencies(ctx context.Context, req *connect.Request[langpb.DependenciesRequest]) (*connect.Response[langpb.DependenciesResponse], error)

	generateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error)
	syncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error)

	build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (chan result.Result[*langpb.BuildEvent], streamCancelFunc, error)
	buildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error)

	kill() error
	cmdErr() <-chan error
}

var _ pluginClient = &pluginClientImpl{}

type pluginClientImpl struct {
	plugin *plugin.Plugin[langconnect.LanguageServiceClient]

	// channel gets closed when the plugin exits
	cmdError chan error
}

func newClientImpl(ctx context.Context, dir, language, name string) (*pluginClientImpl, error) {
	impl := &pluginClientImpl{}
	err := impl.start(ctx, dir, language, name)
	if err != nil {
		return nil, err
	}
	return impl, nil
}

// Start launches the plugin and blocks until the plugin is ready.
func (p *pluginClientImpl) start(ctx context.Context, dir, language, name string) error {
	cmdName := "ftl-language-" + language
	cmdPath, err := exec.LookPath(cmdName)
	if err != nil {
		return fmt.Errorf("failed to find plugin for %s: %w", language, err)
	}

	plugin, cmdCtx, err := plugin.Spawn(ctx,
		log.FromContext(ctx).GetLevel(),
		name,
		dir,
		cmdPath,
		langconnect.NewLanguageServiceClient,
		plugin.WithEnvars("FTL_NAME="+name),
	)
	if err != nil {
		return fmt.Errorf("failed to spawn plugin for %s: %w", name, err)
	}
	p.plugin = plugin

	p.cmdError = make(chan error)
	go func() {
		<-cmdCtx.Done()
		err := cmdCtx.Err()
		if err != nil {
			p.cmdError <- fmt.Errorf("language plugin failed: %w", err)
		} else {
			p.cmdError <- fmt.Errorf("language plugin ended with status 0")
		}
	}()
	return nil
}

func (p *pluginClientImpl) kill() error {
	if err := p.plugin.Cmd.Kill(syscall.SIGINT); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

func (p *pluginClientImpl) cmdErr() <-chan error {
	return p.cmdError
}

func (p *pluginClientImpl) getCreateModuleFlags(ctx context.Context, req *connect.Request[langpb.GetCreateModuleFlagsRequest]) (*connect.Response[langpb.GetCreateModuleFlagsResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.plugin.Client.GetCreateModuleFlags(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) moduleConfigDefaults(ctx context.Context, req *connect.Request[langpb.ModuleConfigDefaultsRequest]) (*connect.Response[langpb.ModuleConfigDefaultsResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.plugin.Client.ModuleConfigDefaults(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) createModule(ctx context.Context, req *connect.Request[langpb.CreateModuleRequest]) (*connect.Response[langpb.CreateModuleResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.plugin.Client.CreateModule(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) getDependencies(ctx context.Context, req *connect.Request[langpb.DependenciesRequest]) (*connect.Response[langpb.DependenciesResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.plugin.Client.GetDependencies(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) generateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.plugin.Client.GenerateStubs(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) syncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.plugin.Client.SyncStubReferences(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (chan result.Result[*langpb.BuildEvent], streamCancelFunc, error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, nil, err
	}
	stream, err := p.plugin.Client.Build(ctx, req)
	if err != nil {
		return nil, nil, err //nolint:wrapcheck
	}

	streamChan := make(chan result.Result[*langpb.BuildEvent], 64)
	go streamToChan(stream, streamChan)

	return streamChan, func() {
		// closing the stream causes the steamToChan goroutine to close the chan
		stream.Close()
	}, nil
}

func streamToChan(stream *connect.ServerStreamForClient[langpb.BuildEvent], ch chan result.Result[*langpb.BuildEvent]) {
	for stream.Receive() {
		ch <- result.From(stream.Msg(), nil)
	}
	if err := stream.Err(); err != nil {
		ch <- result.Err[*langpb.BuildEvent](err)
	}
	close(ch)
}

func (p *pluginClientImpl) buildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.plugin.Client.BuildContextUpdated(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) checkCmdIsAlive() error {
	select {
	case err := <-p.cmdError:
		if err == nil {
			// cmd errored with success or the channel was closed previously
			return ErrPluginNotRunning
		}
		return fmt.Errorf("%w: %w", ErrPluginNotRunning, err)
	default:
		return nil
	}
}
