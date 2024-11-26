package localscaling

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/runner"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/localdebug"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/observability"
)

var _ scaling.RunnerScaling = &localScaling{}

const maxExits = 10

type localScaling struct {
	lock     sync.Mutex
	cacheDir string
	// Module -> Deployments -> info
	runners map[string]map[string]*deploymentInfo
	// Module -> Port
	debugPorts map[string]*localdebug.DebugInfo
	// Module -> Port, most recent runner is present in the map
	portAllocator       *bind.BindAllocator
	controllerAddresses []*url.URL

	prevRunnerSuffix int
	ideSupport       optional.Option[localdebug.IDEIntegration]
	registryConfig   artefacts.RegistryConfig
	enableOtel       bool

	devModeEndpointsUpdates <-chan scaling.DevModeEndpoints
	devModeEndpoints        map[string]*devModeRunner
}

type devModeRunner struct {
	uri url.URL
	// Set to None under mysterious circumstances...
	deploymentKey optional.Option[model.DeploymentKey]
	debugPort     int
}

func (l *localScaling) Start(ctx context.Context, endpoint url.URL, leaser leases.Leaser) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case devEndpoints := <-l.devModeEndpointsUpdates:
				l.lock.Lock()
				l.devModeEndpoints[devEndpoints.Module] = &devModeRunner{
					uri:       devEndpoints.Endpoint,
					debugPort: devEndpoints.DebugPort,
				}
				if ide, ok := l.ideSupport.Get(); ok {
					if devEndpoints.DebugPort != 0 {
						if debug, ok := l.debugPorts[devEndpoints.Module]; ok {
							debug.Port = devEndpoints.DebugPort
						} else {
							l.debugPorts[devEndpoints.Module] = &localdebug.DebugInfo{
								Port:     devEndpoints.DebugPort,
								Language: devEndpoints.Language,
							}
						}
					}
					ide.SyncIDEDebugIntegrations(ctx, l.debugPorts)

				}
				l.lock.Unlock()
			}
		}
	}()
	scaling.BeginGrpcScaling(ctx, endpoint, leaser, l.handleSchemaChange)
	return nil
}

func (l *localScaling) GetEndpointForDeployment(ctx context.Context, module string, deployment string) (optional.Option[url.URL], error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	mod := l.runners[module]
	if mod == nil {
		return optional.None[url.URL](), fmt.Errorf("module %s not found", module)
	}
	dep := mod[deployment]
	if dep == nil {
		return optional.None[url.URL](), fmt.Errorf("deployment %s not found", module)
	}
	if r, ok := dep.runner.Get(); ok {
		return optional.Some(url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%s", r.port),
		}), nil
	}
	return optional.None[url.URL](), nil
}

type deploymentInfo struct {
	runner   optional.Option[runnerInfo]
	module   string
	replicas int32
	key      model.DeploymentKey
	language string
	exits    int
}
type runnerInfo struct {
	cancelFunc context.CancelFunc
	port       string
}

func NewLocalScaling(portAllocator *bind.BindAllocator, controllerAddresses []*url.URL, configPath string, enableIDEIntegration bool, registryConfig artefacts.RegistryConfig, enableOtel bool, devModeEndpoints <-chan scaling.DevModeEndpoints) (scaling.RunnerScaling, error) {

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	local := localScaling{
		lock:                    sync.Mutex{},
		cacheDir:                cacheDir,
		runners:                 map[string]map[string]*deploymentInfo{},
		portAllocator:           portAllocator,
		controllerAddresses:     controllerAddresses,
		prevRunnerSuffix:        -1,
		debugPorts:              map[string]*localdebug.DebugInfo{},
		registryConfig:          registryConfig,
		enableOtel:              enableOtel,
		devModeEndpointsUpdates: devModeEndpoints,
		devModeEndpoints:        map[string]*devModeRunner{},
	}
	if enableIDEIntegration && configPath != "" {
		local.ideSupport = optional.Ptr(localdebug.NewIDEIntegration(configPath))
	}

	return &local, nil
}

func (l *localScaling) handleSchemaChange(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
	if msg.DeploymentKey == nil {
		// Builtins don't have deployments
		return nil
	}
	deploymentKey, err := model.ParseDeploymentKey(msg.GetDeploymentKey())
	if err != nil {
		return fmt.Errorf("failed to parse deployment key: %w", err)
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	logger := log.FromContext(ctx).Scope("localScaling").Module(msg.ModuleName)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Handling schema change for %s", deploymentKey)
	moduleDeployments := l.runners[msg.ModuleName]
	if moduleDeployments == nil {
		moduleDeployments = map[string]*deploymentInfo{}
		l.runners[msg.ModuleName] = moduleDeployments
	}
	deploymentRunners := moduleDeployments[deploymentKey.String()]
	if deploymentRunners == nil {
		deploymentRunners = &deploymentInfo{runner: optional.None[runnerInfo](), key: deploymentKey, module: msg.ModuleName, language: msg.Schema.Runtime.Language}
		moduleDeployments[deploymentKey.String()] = deploymentRunners
	}

	switch msg.ChangeType {
	case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:
		deploymentRunners.replicas = msg.Schema.Runtime.MinReplicas
	case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
		deploymentRunners.replicas = 0
		delete(moduleDeployments, deploymentKey.String())
	}
	return l.reconcileRunners(ctx, deploymentRunners)
}

func (l *localScaling) reconcileRunners(ctx context.Context, deploymentRunners *deploymentInfo) error {
	// Must be called under lock
	logger := log.FromContext(ctx)
	if deploymentRunners.replicas > 0 && !deploymentRunners.runner.Ok() && deploymentRunners.exits < maxExits {
		if err := l.startRunner(ctx, deploymentRunners.key, deploymentRunners); err != nil {
			logger.Errorf(err, "Failed to start runner")
			return err
		}
	} else if deploymentRunners.replicas == 0 && deploymentRunners.runner.Ok() {
		go func() {
			// Nasty hack, we want all the controllers to have updated their route tables before we kill the runner
			// so we add a slight delay here
			time.Sleep(time.Second * 2)
			l.lock.Lock()
			defer l.lock.Unlock()
			if r, ok := deploymentRunners.runner.Get(); ok {
				r.cancelFunc()
			}
		}()
		deploymentRunners.runner = optional.None[runnerInfo]()
	}
	return nil
}

func (l *localScaling) startRunner(ctx context.Context, deploymentKey model.DeploymentKey, info *deploymentInfo) error {
	select {
	case <-ctx.Done():
		// In some cases this gets called with an expired context, generally after the lease is released
		// We don't want to start a runner in that case
		return nil
	default:
	}

	devEndpoint := l.devModeEndpoints[info.module]
	devURI := optional.None[url.URL]()
	debugPort := 0
	if devEndpoint != nil {
		devURI = optional.Some(devEndpoint.uri)
		if devKey, ok := devEndpoint.deploymentKey.Get(); ok && devKey.Equal(deploymentKey) {
			// Already running, don't start another
			return nil
		}
		devEndpoint.deploymentKey = optional.Some(deploymentKey)
		debugPort = devEndpoint.debugPort
	} else if ide, ok := l.ideSupport.Get(); ok {
		var debug *localdebug.DebugInfo
		debugBind, err := l.portAllocator.NextPort()
		if err != nil {
			return fmt.Errorf("failed to start runner: %w", err)
		}
		debug = &localdebug.DebugInfo{
			Language: info.language,
			Port:     debugBind,
		}
		l.debugPorts[info.module] = debug
		ide.SyncIDEDebugIntegrations(ctx, l.debugPorts)
		debugPort = debug.Port
	}
	controllerEndpoint := l.controllerAddresses[len(l.runners)%len(l.controllerAddresses)]
	logger := log.FromContext(ctx)

	bind, err := l.portAllocator.Next()
	if err != nil {
		return fmt.Errorf("failed to start runner: %w", err)
	}

	keySuffix := l.prevRunnerSuffix + 1
	l.prevRunnerSuffix = keySuffix

	config := runner.Config{
		Bind:               bind,
		ControllerEndpoint: controllerEndpoint,
		Key:                model.NewLocalRunnerKey(keySuffix),
		Deployment:         deploymentKey,
		DebugPort:          debugPort,
		Registry:           l.registryConfig,
		ObservabilityConfig: observability.Config{
			ExportOTEL: observability.ExportOTELFlag(l.enableOtel),
		},
		DevEndpoint: devURI,
	}

	simpleName := fmt.Sprintf("runner%d", keySuffix)
	if err := kong.ApplyDefaults(&config, kong.Vars{
		"deploymentdir": filepath.Join(l.cacheDir, "ftl-runner", simpleName, "deployments"),
		"language":      "go,kotlin,java",
	}); err != nil {
		return fmt.Errorf("failed to apply defaults: %w", err)
	}
	config.HeartbeatPeriod = time.Second
	config.HeartbeatJitter = time.Millisecond * 100

	runnerCtx := log.ContextWithLogger(ctx, logger.Scope(simpleName).Module(info.module))

	runnerCtx, cancel := context.WithCancel(runnerCtx)
	info.runner = optional.Some(runnerInfo{cancelFunc: cancel, port: bind.Port()})

	go func() {
		logger.Debugf("Starting runner: %s", config.Key)
		err := runner.Start(runnerCtx, config)
		l.lock.Lock()
		defer l.lock.Unlock()
		if devEndpoint != nil {
			devEndpoint.deploymentKey = optional.None[model.DeploymentKey]()
		}
		// Don't count context.Canceled as an a restart error
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf(err, "Runner failed: %s", err)
			info.exits++
		}
		if info.exits >= maxExits {
			logger.Errorf(fmt.Errorf("too many restarts"), "Runner failed too many times, not restarting")
		}
		info.runner = optional.None[runnerInfo]()
		err = l.reconcileRunners(ctx, info)
		if err != nil {
			logger.Errorf(err, "Failed to reconcile runners")
		}
	}()
	return nil
}
