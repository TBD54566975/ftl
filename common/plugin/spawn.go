package plugin

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/common/exec"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/rpc"
	"github.com/TBD54566975/ftl/common/socket"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

// PingableClient is a gRPC client that can be pinged.
type PingableClient interface {
	Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

type pluginOptions struct {
	envars            []string
	additionalClients []func(baseURL string, opts ...connect.ClientOption)
	startTimeout      time.Duration
}

// Option used when creating a plugin.
type Option func(*pluginOptions) error

// WithEnvars sets the environment variables to pass to the plugin.
func WithEnvars(envars ...string) Option {
	return func(po *pluginOptions) error {
		po.envars = append(po.envars, envars...)
		return nil
	}
}

// WithStartTimeout sets the timeout for the language-specific drive plugin to start.
func WithStartTimeout(timeout time.Duration) Option {
	return func(po *pluginOptions) error {
		po.startTimeout = timeout
		return nil
	}
}

// WithExtraClient connects to an additional gRPC service in the same plugin.
//
// The client instance is written to "out".
func WithExtraClient[Client PingableClient](out *Client, makeClient rpc.ClientFactory[Client]) Option {
	return func(po *pluginOptions) error {
		po.additionalClients = append(po.additionalClients, func(baseURL string, opts ...connect.ClientOption) {
			*out = rpc.Dial(makeClient, baseURL, opts...)
		})
		return nil
	}
}

type Plugin[Client PingableClient] struct {
	Cmd    *exec.Cmd
	Client Client
}

// Spawn a new sub-process plugin.
//
// Plugins are gRPC servers that listen on a socket passed in an envar.
//
// If the subprocess is a Go plugin, it should call [Start] to start the gRPC
// server.
//
// "cmdCtx" will be cancelled when the plugin stops.
//
// The envars passed to the plugin are:
//
//	FTL_PLUGIN_ENDPOINT - the endpoint URI to listen on
//	FTL_WORKING_DIR - the path to a working directory that the plugin can write state to, if required.
func Spawn[Client PingableClient](
	ctx context.Context,
	name, dir, exe string,
	makeClient rpc.ClientFactory[Client],
	options ...Option,
) (
	plugin *Plugin[Client],
	// "cmdCtx" will be cancelled when the plugin process stops.
	cmdCtx context.Context,
	err error,
) {
	logger := log.FromContext(ctx).Sub(name, log.Default)

	opts := pluginOptions{
		startTimeout: time.Second * 30,
	}
	for _, opt := range options {
		if err = opt(&opts); err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}
	workingDir := filepath.Join(dir, ".ftl")

	// Clean up previous process.
	pidFile := filepath.Join(workingDir, filepath.Base(exe)+".pid")
	err = cleanup(logger, pidFile)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Find a free port.
	addr, err := allocatePort()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Start the plugin process.
	pluginSocket := socket.Socket{Network: "tcp", Addr: addr.String()}
	logger.Debugf("Spawning plugin on %s", pluginSocket)
	cmd := exec.Command(ctx, dir, exe)

	// Send the plugin's stdout to the logger.
	cmd.Stderr = nil
	pipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	cmd.Env = append(cmd.Env, "FTL_PLUGIN_ENDPOINT="+pluginSocket.String())
	cmd.Env = append(cmd.Env, "FTL_WORKING_DIR="+workingDir)
	cmd.Env = append(cmd.Env, opts.envars...)
	if err = cmd.Start(); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	// Cancel the context if the command exits - this will terminate the Dial immediately.
	var cancelWithCause context.CancelCauseFunc
	cmdCtx, cancelWithCause = context.WithCancelCause(ctx)
	go func() { cancelWithCause(cmd.Wait()) }()

	go func() {
		err := log.JSONStreamer(pipe, logger, log.Info)
		if err != nil {
			logger.Errorf(err, "Error streaming plugin logs.")
		}
	}()

	defer func() {
		if err != nil {
			logger.Warnf("Plugin failed to start, terminating pid %d", cmd.Process.Pid)
			_ = cmd.Kill(syscall.SIGTERM)
		}
	}()

	// Write the PID file.
	err = ioutil.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0600)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	dialCtx, cancel := context.WithTimeout(ctx, opts.startTimeout)
	defer cancel()

	// Wait for the plugin to start.
	client := rpc.Dial(makeClient, pluginSocket.URL())
	pingErr := make(chan error)
	go func() {
		defer close(pingErr)
		for {
			select {
			case <-dialCtx.Done():
				return
			default:
			}
			_, err := client.Ping(dialCtx, connect.NewRequest(&ftlv1.PingRequest{}))
			if err != nil {
				logger.Tracef("Plugin ping failed: %+v", err)
				time.Sleep(time.Millisecond * 100)
				continue
			}
			pingErr <- err
		}
	}()

	select {
	case <-dialCtx.Done():
		return nil, nil, errors.Wrap(dialCtx.Err(), "plugin timed out while starting")

	case <-cmdCtx.Done():
		return nil, nil, errors.Wrap(cmdCtx.Err(), "plugin process died")

	case err := <-pingErr:
		if err != nil {
			return nil, nil, errors.Wrap(err, "plugin failed to respond to ping")
		}
	}

	for _, makeClient := range opts.additionalClients {
		makeClient(pluginSocket.URL())
	}

	logger.Infof("Online")
	plugin = &Plugin[Client]{Cmd: cmd, Client: client}
	return plugin, cmdCtx, nil
}
