package provisioner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"connectrpc.com/connect"
	"github.com/BurntSushi/toml"
	"github.com/alecthomas/kong"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/sync/errgroup"

	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	ftlv1connect "github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner/scaling"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

// CommonProvisionerConfig is shared config between the production controller and development server.
type CommonProvisionerConfig struct {
	PluginConfigFile *os.File `name:"provisioner-plugin-config" help:"Path to the plugin configuration file." env:"FTL_PROVISIONER_PLUGIN_CONFIG_FILE"`
}

type Config struct {
	Bind               *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8893" env:"FTL_BIND"`
	ControllerEndpoint *url.URL `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	CommonProvisionerConfig
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c); err != nil {
		panic(err)
	}
}

type Service struct {
	currentModules   *xsync.MapOf[string, *schema.Module]
	controllerClient ftlv1connect.ControllerServiceClient
	registry         *ProvisionerRegistry
}

var _ provisionerconnect.ProvisionerServiceHandler = (*Service)(nil)

func New(
	ctx context.Context,
	config Config,
	controllerClient ftlv1connect.ControllerServiceClient,
	registry *ProvisionerRegistry,
) (*Service, error) {
	return &Service{
		controllerClient: controllerClient,
		currentModules:   xsync.NewMapOf[string, *schema.Module](),
		registry:         registry,
	}, nil
}

func (s *Service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (s *Service) CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	logger := log.FromContext(ctx)
	// TODO: Block deployments to make sure only one module is modified at a time
	moduleName := req.Msg.Schema.Name

	existingModule, _ := s.currentModules.Load(moduleName)
	desiredModule, err := schema.ModuleFromProto(req.Msg.Schema)
	if err != nil {
		return nil, fmt.Errorf("error converting module to schema: %w", err)
	}

	if existingModule != nil {
		syncExistingRuntimes(existingModule, desiredModule)
	}

	// Place artefacts to the metdata
	desiredModule.Metadata = slices.Filter(desiredModule.Metadata, func(m schema.Metadata) bool {
		_, ok := m.(*schema.MetadataArtefact)
		return !ok
	})
	for _, artefact := range req.Msg.Artefacts {
		desiredModule.Metadata = append(desiredModule.Metadata, &schema.MetadataArtefact{
			Path:       artefact.Path,
			Digest:     artefact.Digest,
			Executable: artefact.Executable,
		})
	}

	deployment := s.registry.CreateDeployment(ctx, desiredModule, existingModule)
	running := true
	logger.Debugf("Running deployment for module %s", moduleName)
	for running {
		running, err = deployment.Progress(ctx)
		if err != nil {
			// TODO: Deal with failed deployments
			return nil, fmt.Errorf("error running a provisioner: %w", err)
		}
	}
	logger.Debugf("Finished deployment for module %s", moduleName)

	deploymentKey := deployment.Module.Runtime.Deployment.DeploymentKey
	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{
		DeploymentKey: deploymentKey,
	}), nil
}

// Start the Provisioner. Blocks until the context is cancelled.
func Start(
	ctx context.Context,
	config Config,
	registry *ProvisionerRegistry,
	controllerClient ftlv1connect.ControllerServiceClient,
) error {
	config.SetDefaults()

	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL provisioner")

	svc, err := New(ctx, config, controllerClient, registry)
	if err != nil {
		return err
	}
	logger.Debugf("Provisioner available at: %s", config.Bind)
	logger.Debugf("Using FTL endpoint: %s", config.ControllerEndpoint)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return rpc.Serve(ctx, config.Bind,
			rpc.GRPC(provisionerconnect.NewProvisionerServiceHandler, svc),
			rpc.PProf(),
		)
	})
	if err := g.Wait(); err != nil {
		return fmt.Errorf("error waiting for rpc.Serve: %w", err)
	}
	return nil
}

func RegistryFromConfigFile(ctx context.Context, file *os.File, controller ftlv1connect.ControllerServiceClient, scaling scaling.RunnerScaling) (*ProvisionerRegistry, error) {
	config := provisionerPluginConfig{}
	bytes, err := io.ReadAll(bufio.NewReader(file))
	if err != nil {
		return nil, fmt.Errorf("error reading plugin configuration from %s: %w", file.Name(), err)
	}
	if err := toml.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("error parsing plugin configuration: %w", err)
	}

	registry, err := registryFromConfig(ctx, &config, controller, scaling)
	if err != nil {
		return nil, fmt.Errorf("error creating provisioner registry: %w", err)
	}

	return registry, nil
}

// Deployment client calls to ftl-controller

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	resp, err := s.controllerClient.GetArtefactDiffs(ctx, req)

	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *Service) ReplaceDeploy(ctx context.Context, req *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error) {
	resp, err := s.controllerClient.ReplaceDeploy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *Service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	resp, err := s.controllerClient.Status(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *Service) UpdateDeploy(ctx context.Context, req *connect.Request[ftlv1.UpdateDeployRequest]) (*connect.Response[ftlv1.UpdateDeployResponse], error) {
	resp, err := s.controllerClient.UpdateDeploy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	resp, err := s.controllerClient.UploadArtefact(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func syncExistingRuntimes(existingModule, desiredModule *schema.Module) {
	existingResources := schema.GetProvisioned(existingModule)
	desiredResources := schema.GetProvisioned(desiredModule)

	for id, desired := range desiredResources {
		if existing, ok := existingResources[id]; ok {
			switch desired := desired.(type) {
			case *schema.Database:
				if existing, ok := existing.(*schema.Database); ok {
					desired.Runtime = reflect.DeepCopy(existing.Runtime)
				}
			case *schema.Topic:
				if existing, ok := existing.(*schema.Topic); ok {
					desired.Runtime = reflect.DeepCopy(existing.Runtime)
				}
			case *schema.Verb:
				if existing, ok := existing.(*schema.Verb); ok {
					desired.Runtime = reflect.DeepCopy(existing.Runtime)
				}
			case *schema.Module:
				if existing, ok := existing.(*schema.Module); ok {
					desired.Runtime = reflect.DeepCopy(existing.Runtime)
				}
			}
		}
	}
}
