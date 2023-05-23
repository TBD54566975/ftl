// Package runnergo contains a server that implements the RunnerService and
// proxies VerbService requests to user code.
package runnergo

import (
	context "context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/common/download"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/common/rpc"
	"github.com/TBD54566975/ftl/common/server"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	sdkgo "github.com/TBD54566975/ftl/sdk-go"
)

type Config struct {
	Bind          *url.URL `help:"Socket to bind to." default:"http://localhost:8893"`
	FTLEndpoint   *url.URL `help:"FTL endpoint." env:"FTL_ENDPOINT" default:"http://localhost:8892"`
	DeploymentDir string   `help:"Directory to store deployments in." default:"${deploymentdir}"`
}

func Start(ctx context.Context, config Config) error {
	ctx = sdkgo.ContextWithClient(ctx, config.FTLEndpoint)
	logger := log.FromContext(ctx)
	logger.Infof("Starting FTL runner")
	logger.Infof("Deployment directory: %s", config.DeploymentDir)
	err := os.MkdirAll(config.DeploymentDir, 0700)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment directory")
	}
	logger.Infof("Using FTL endpoint: %s", config.FTLEndpoint)
	logger.Infof("Listening on %s", config.Bind)

	backplaneClient := rpc.Dial(ftlv1connect.NewBackplaneServiceClient, config.FTLEndpoint.String())

	svc := &Service{
		backplaneClient: backplaneClient,
		deploymentDir:   config.DeploymentDir,
		ftlEndpoint:     config.FTLEndpoint,
	}
	reflector := grpcreflect.NewStaticReflector(ftlv1connect.RunnerServiceName, ftlv1connect.VerbServiceName)
	return server.Serve(ctx, config.Bind,
		server.Route("/"+ftlv1connect.VerbServiceName+"/", svc), // The Runner proxies all verbs to the deployment.
		server.GRPC(ftlv1connect.NewRunnerServiceHandler, svc),
		server.Route(grpcreflect.NewHandlerV1(reflector)),
		server.Route(grpcreflect.NewHandlerV1Alpha(reflector)),
	)
}

var _ ftlv1connect.RunnerServiceHandler = (*Service)(nil)
var _ http.Handler = (*Service)(nil)

type pluginProxy struct {
	plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]
	proxy  *httputil.ReverseProxy
}

type Service struct {
	// We use double-checked locking around the atomic so that the read fast-path is lock-free.
	lock       sync.Mutex
	deployment atomic.Value[*pluginProxy]

	backplaneClient ftlv1connect.BackplaneServiceClient
	deploymentDir   string
	ftlEndpoint     *url.URL
}

// ServeHTTP proxies through to the deployment.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if deployment := s.deployment.Load(); deployment != nil {
		deployment.proxy.ServeHTTP(w, r)
	} else {
		http.Error(w, "503 No deployment", http.StatusServiceUnavailable)
	}
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) Deploy(ctx context.Context, req *connect.Request[ftlv1.DeployRequest]) (*connect.Response[ftlv1.DeployResponse], error) {
	id, err := uuid.Parse(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}

	// Double-checked lock.
	if s.deployment.Load() != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("already deployed"))
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load() != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("already deployed"))
	}

	gdResp, err := s.backplaneClient.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{DeploymentKey: req.Msg.DeploymentKey}))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = os.Mkdir(filepath.Join(s.deploymentDir, id.String()), 0700)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create deployment directory")
	}
	err = download.Artefacts(ctx, s.backplaneClient, id, s.deploymentDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download artefacts")
	}
	deployment, _, err := plugin.Spawn(
		ctx,
		gdResp.Msg.Schema.Name,
		filepath.Join(s.deploymentDir, id.String()),
		"./main",
		ftlv1connect.NewVerbServiceClient,
		plugin.WithEnvars("FTL_ENDPOINT="+s.ftlEndpoint.String()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to spawn plugin")
	}
	s.deployment.Store(s.makePluginProxy(deployment))
	return connect.NewResponse(&ftlv1.DeployResponse{}), nil
}

func (s *Service) makePluginProxy(plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]) *pluginProxy {
	return &pluginProxy{
		plugin: plugin,
		proxy:  httputil.NewSingleHostReverseProxy(plugin.Endpoint),
	}
}
