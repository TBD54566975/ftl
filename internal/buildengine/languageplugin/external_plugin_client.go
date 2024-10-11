package languageplugin

import (
	"context"
	"fmt"
	"net/url"
	"syscall"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/either"
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

	build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (chan either.Either[*langpb.BuildEvent, error], streamCancelFunc, error)
	buildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error)

	kill() error
}

var _ externalPluginClient = &externalPluginImpl{}

type externalPluginImpl struct {
	cmd    *exec.Cmd
	client langconnect.LanguageServiceClient
}

func newExternalPluginImpl(ctx context.Context, bind *url.URL, language string) (*externalPluginImpl, error) {
	impl := &externalPluginImpl{
		client: rpc.Dial(langconnect.NewLanguageServiceClient, bind.String(), log.Error),
	}
	err := impl.start(ctx, bind, language)
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}
	return impl, nil
}

// Start launches the plugin and blocks until the plugin is ready.
func (p *externalPluginImpl) start(ctx context.Context, bind *url.URL, language string) error {
	cmdName := "ftl-language-" + language
	p.cmd = exec.Command(ctx, log.Debug, ".", cmdName, "--bind", bind.String())
	_, err := exec.LookPath(cmdName)
	if err != nil {
		return fmt.Errorf("failed to find plugin for %s: %w", language, err)
	}

	runCtx, cancel := context.WithCancel(ctx)

	cmdErr := make(chan error)
	pingErr := make(chan error)

	// run the plugin and wait for it to finish executing
	go func() {
		err := p.cmd.RunBuffered(runCtx)
		if err != nil {
			cmdErr <- fmt.Errorf("language plugin failed: %w", err)
			cancel()
		}
		close(cmdErr)
	}()
	go func() {
		// wait for the plugin to be ready
		if err := p.ping(runCtx); err != nil {
			cancel()
			pingErr <- fmt.Errorf("failed to ping plugin")
		}
		close(pingErr)
	}()

	select {
	case err := <-cmdErr:
		return err
	case err := <-pingErr:
		if err != nil {
			return nil
		}
		return fmt.Errorf("failed to start plugin: %w", err)
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
	if err := p.cmd.Kill(syscall.SIGINT); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

func (p *externalPluginImpl) getCreateModuleFlags(ctx context.Context, req *connect.Request[langpb.GetCreateModuleFlagsRequest]) (*connect.Response[langpb.GetCreateModuleFlagsResponse], error) {
	resp, err := p.client.GetCreateModuleFlags(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) moduleConfigDefaults(ctx context.Context, req *connect.Request[langpb.ModuleConfigDefaultsRequest]) (*connect.Response[langpb.ModuleConfigDefaultsResponse], error) {
	resp, err := p.client.ModuleConfigDefaults(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) createModule(ctx context.Context, req *connect.Request[langpb.CreateModuleRequest]) (*connect.Response[langpb.CreateModuleResponse], error) {
	resp, err := p.client.CreateModule(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) getDependencies(ctx context.Context, req *connect.Request[langpb.DependenciesRequest]) (*connect.Response[langpb.DependenciesResponse], error) {
	resp, err := p.client.GetDependencies(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}

func (p *externalPluginImpl) build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (chan either.Either[*langpb.BuildEvent, error], streamCancelFunc, error) {
	stream, err := p.client.Build(ctx, req)
	if err != nil {
		return nil, nil, err //nolint:wrapcheck
	}

	streamChan := make(chan either.Either[*langpb.BuildEvent, error], 64)
	go streamToChan(stream, streamChan)

	return streamChan, func() {
		stream.Close()
		close(streamChan)
	}, nil
}

func streamToChan(stream *connect.ServerStreamForClient[langpb.BuildEvent], ch chan either.Either[*langpb.BuildEvent, error]) {
	for stream.Receive() {
		ch <- either.LeftOf[error](stream.Msg())
	}
	if err := stream.Err(); err != nil {
		ch <- either.RightOf[*langpb.BuildEvent](err)
	}
	close(ch)
}

func (p *externalPluginImpl) buildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error) {
	resp, err := p.client.BuildContextUpdated(ctx, req)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return resp, nil
}
