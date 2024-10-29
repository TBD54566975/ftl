package languageplugin

import (
	"context"
	"fmt"
	"net/url"
	"syscall"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/result"
	"github.com/jpillora/backoff"

	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	langconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language/languagepbconnect"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type streamCancelFunc func()

type externalPluginClient interface {
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

var _ externalPluginClient = &externalPluginImpl{}

type externalPluginImpl struct {
	cmd    *exec.Cmd
	client langconnect.LanguageServiceClient

	// channel gets closed when the plugin exits
	cmdError chan error
}

func newExternalPluginImpl(ctx context.Context, bind *url.URL, language, name string) (*externalPluginImpl, error) {
	impl := &externalPluginImpl{
		client: rpc.Dial(langconnect.NewLanguageServiceClient, bind.String(), log.Error),
	}
	err := impl.start(ctx, bind, language, name)
	if err != nil {
		return nil, err
	}
	return impl, nil
}

// Start launches the plugin and blocks until the plugin is ready.
func (p *externalPluginImpl) start(ctx context.Context, bind *url.URL, language, name string) error {
	logger := log.FromContext(ctx).Scope(name)

	cmdName := "ftl-language-" + language
	p.cmd = exec.Command(ctx, log.Debug, ".", cmdName, "--bind", bind.String())
	p.cmd.Env = append(p.cmd.Env, "FTL_NAME="+name)
	_, err := exec.LookPath(cmdName)
	if err != nil {
		return fmt.Errorf("failed to find plugin for %s: %w", language, err)
	}

	// Send the plugin's stderr to the logger.
	p.cmd.Stderr = nil
	pipe, err := p.cmd.StderrPipe()
	if err != nil {
		return err
	}
	go func() {
		err := log.JSONStreamer(pipe, logger, log.Error)
		if err != nil {
			logger.Errorf(err, "Error streaming plugin logs.")
		}
	}()

	runCtx, cancel := context.WithCancel(ctx)

	p.cmdError = make(chan error)
	pingErr := make(chan error)

	// run the plugin and wait for it to finish executing
	go func() {
		err := p.cmd.Run()
		if err != nil {
			p.cmdError <- fmt.Errorf("language plugin failed: %w", err)
		} else {
			p.cmdError <- fmt.Errorf("language plugin ended with status 0")
		}
		cancel()
		close(p.cmdError)
	}()
	go func() {
		// wait for the plugin to be ready
		if err := p.ping(runCtx); err != nil {
			cancel()
			pingErr <- fmt.Errorf("failed to ping plugin")
		}
		close(pingErr)
	}()

	// Wait for ping result, or for the plugin to exit. Which ever happens first.
	select {
	case err := <-p.cmdError:
		if err != nil {
			return err
		}
		return fmt.Errorf("plugin exited with status 0 before ping was registered")
	case err := <-pingErr:
		if err != nil {
			return fmt.Errorf("failed to start plugin: %w", err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("failed to start plugin: %w", ctx.Err())
	}
}

func (p *externalPluginImpl) ping(ctx context.Context) error {
	retry := backoff.Backoff{}
	heartbeatCtx, cancel := context.WithTimeout(ctx, launchTimeout)
	defer cancel()
	err := rpc.Wait(heartbeatCtx, retry, p.client)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to connect to runner: %w", err))
	}
	return nil
}

func (p *externalPluginImpl) kill() error {
	// close cmdErr so that we don't publish cmd error.
	close(p.cmdError)
	if err := p.cmd.Kill(syscall.SIGINT); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

func (p *externalPluginImpl) cmdErr() <-chan error {
	return p.cmdError
}

func (p *externalPluginImpl) getCreateModuleFlags(ctx context.Context, req *connect.Request[langpb.GetCreateModuleFlagsRequest]) (*connect.Response[langpb.GetCreateModuleFlagsResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.client.GetCreateModuleFlags(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) moduleConfigDefaults(ctx context.Context, req *connect.Request[langpb.ModuleConfigDefaultsRequest]) (*connect.Response[langpb.ModuleConfigDefaultsResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.client.ModuleConfigDefaults(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) createModule(ctx context.Context, req *connect.Request[langpb.CreateModuleRequest]) (*connect.Response[langpb.CreateModuleResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.client.CreateModule(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) getDependencies(ctx context.Context, req *connect.Request[langpb.DependenciesRequest]) (*connect.Response[langpb.DependenciesResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.client.GetDependencies(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) generateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.client.GenerateStubs(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) syncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.client.SyncStubReferences(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (chan result.Result[*langpb.BuildEvent], streamCancelFunc, error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, nil, err
	}
	stream, err := p.client.Build(ctx, req)
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

func (p *externalPluginImpl) buildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, err
	}
	resp, err := p.client.BuildContextUpdated(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) checkCmdIsAlive() error {
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
