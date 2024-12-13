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
	"github.com/TBD54566975/ftl/backend/provisioner/scaling"
	"github.com/TBD54566975/ftl/backend/runner"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/internal/dev"
	"github.com/TBD54566975/ftl/internal/localdebug"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

var _ scaling.RunnerScaling = &localScaling{}

const maxExits = 10

type localScaling struct {
	ctx      context.Context
	lock     sync.Mutex
	cacheDir string
	// Module -> Deployments -> info
	runners map[string]map[string]*deploymentInfo
	// Module -> Port
	debugPorts          map[string]*localdebug.DebugInfo
	controllerAddresses []*url.URL
	leaseAddress        *url.URL

	prevRunnerSuffix int
	ideSupport       optional.Option[localdebug.IDEIntegration]
	storage          *artefacts.OCIArtefactService
	enableOtel       bool

	devModeEndpointsUpdates <-chan dev.LocalEndpoint
	devModeEndpoints        map[string]*devModeRunner
}

func (l *localScaling) StartDeployment(ctx context.Context, module string, deployment string, sch *schema.Module, hasCron bool, hasIngress bool) error {
	if sch.Runtime == nil {
		return nil
	}
	return l.setReplicas(module, deployment, sch.Runtime.Base.Language, 1)
}

func (l *localScaling) setReplicas(module string, deployment string, language string, replicas int32) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	deploymentKey, err := model.ParseDeploymentKey(deployment)
	if err != nil {
		return fmt.Errorf("failed to parse deployment key: %w", err)
	}
	ctx := l.ctx
	logger := log.FromContext(ctx).Scope("localScaling").Module(module)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Starting deployment for %s", deployment)
	moduleDeployments := l.runners[module]
	if moduleDeployments == nil {
		moduleDeployments = map[string]*deploymentInfo{}
		l.runners[module] = moduleDeployments
	}
	deploymentRunners := moduleDeployments[deployment]
	if deploymentRunners == nil {
		deploymentRunners = &deploymentInfo{runner: optional.None[runnerInfo](), key: deploymentKey, module: module, language: language}
		moduleDeployments[deployment] = deploymentRunners
	}
	deploymentRunners.replicas = replicas

	return l.reconcileRunners(ctx, deploymentRunners)
}

func (l *localScaling) TerminatePreviousDeployments(ctx context.Context, module string, deployment string) ([]string, error) {
	logger := log.FromContext(ctx)
	var ret []string
	// So hacky, all this needs to change when the provisioner is a proper schema observer
	logger.Debugf("Terminating previous deployments for %s", deployment)
	for dep := range l.runners[module] {
		if dep != deployment {
			ret = append(ret, dep)
			logger.Debugf("Terminating deployment %s", dep)
			if err := l.setReplicas(module, dep, "", 0); err != nil {
				return nil, err
			}
		}
	}
	return ret, nil
}

type devModeRunner struct {
	uri url.URL
	// The deployment key of the deployment that is currently running
	deploymentKey  optional.Option[model.DeploymentKey]
	debugPort      int
	runnerInfoFile optional.Option[string]
}

func (l *localScaling) Start(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case devEndpoints := <-l.devModeEndpointsUpdates:
				l.lock.Lock()
				l.updateDevModeEndpoint(ctx, devEndpoints)
				l.lock.Unlock()
			}
		}
	}()
	return nil
}

// updateDevModeEndpoint updates the dev mode endpoint for a module
// Must be called under lock
func (l *localScaling) updateDevModeEndpoint(ctx context.Context, devEndpoints dev.LocalEndpoint) {
	l.devModeEndpoints[devEndpoints.Module] = &devModeRunner{
		uri:            devEndpoints.Endpoint,
		debugPort:      devEndpoints.DebugPort,
		runnerInfoFile: devEndpoints.RunnerInfoFile,
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
			Host:   fmt.Sprintf("%s:%d", r.host, r.port),
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
	port       int
	host       string
}

func NewLocalScaling(
	ctx context.Context,
	controllerAddresses []*url.URL,
	leaseAddress *url.URL,
	configPath string,
	enableIDEIntegration bool,
	storage *artefacts.OCIArtefactService,
	enableOtel bool,
	devModeEndpoints <-chan dev.LocalEndpoint,
) (scaling.RunnerScaling, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	local := localScaling{
		ctx:                     ctx,
		lock:                    sync.Mutex{},
		cacheDir:                cacheDir,
		runners:                 map[string]map[string]*deploymentInfo{},
		controllerAddresses:     controllerAddresses,
		leaseAddress:            leaseAddress,
		prevRunnerSuffix:        -1,
		debugPorts:              map[string]*localdebug.DebugInfo{},
		storage:                 storage,
		enableOtel:              enableOtel,
		devModeEndpointsUpdates: devModeEndpoints,
		devModeEndpoints:        map[string]*devModeRunner{},
	}
	if enableIDEIntegration && configPath != "" {
		local.ideSupport = optional.Ptr(localdebug.NewIDEIntegration(configPath))
	}

	return &local, nil
}

func (l *localScaling) reconcileRunners(ctx context.Context, deploymentRunners *deploymentInfo) error {
	// Must be called under lock

	// First make sure we have all endpoint updates
	for {
		select {
		case devEndpoints := <-l.devModeEndpointsUpdates:
			l.updateDevModeEndpoint(ctx, devEndpoints)
			continue
		default:
		}
		break
	}

	logger := log.FromContext(ctx)
	if deploymentRunners.replicas > 0 && !deploymentRunners.runner.Ok() && deploymentRunners.exits < maxExits {
		if err := l.startRunner(ctx, deploymentRunners.key, deploymentRunners); err != nil {
			logger.Errorf(err, "Failed to start runner")
			return err
		}
	} else if runner, ok := deploymentRunners.runner.Get(); deploymentRunners.replicas == 0 && ok {
		go func() {
			// Nasty hack, we want all the controllers to have updated their route tables before we kill the runner
			// so we add a slight delay here
			time.Sleep(time.Second * 5)
			l.lock.Lock()
			defer l.lock.Unlock()
			runner.cancelFunc()
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
	devRunnerInfoFile := optional.None[string]()
	debugPort := 0
	if devEndpoint != nil {
		devURI = optional.Some(devEndpoint.uri)
		devRunnerInfoFile = devEndpoint.runnerInfoFile
		if devKey, ok := devEndpoint.deploymentKey.Get(); ok && devKey.Equal(deploymentKey) {
			// Already running, don't start another
			return nil
		}
		devEndpoint.deploymentKey = optional.Some(deploymentKey)
		debugPort = devEndpoint.debugPort
	} else if ide, ok := l.ideSupport.Get(); ok {
		var debug *localdebug.DebugInfo
		debugBind, err := plugin.AllocatePort()
		if err != nil {
			return fmt.Errorf("failed to start runner: %w", err)
		}
		debug = &localdebug.DebugInfo{
			Language: info.language,
			Port:     debugBind.Port,
		}
		l.debugPorts[info.module] = debug
		ide.SyncIDEDebugIntegrations(ctx, l.debugPorts)
		debugPort = debug.Port
	}
	controllerEndpoint := l.controllerAddresses[len(l.runners)%len(l.controllerAddresses)]
	logger := log.FromContext(ctx)

	bind, err := plugin.AllocatePort()
	if err != nil {
		return fmt.Errorf("failed to start runner: %w", err)
	}

	keySuffix := l.prevRunnerSuffix + 1
	l.prevRunnerSuffix = keySuffix

	bindURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", bind.Port))
	if err != nil {
		return fmt.Errorf("failed to start runner: %w", err)
	}
	config := runner.Config{
		Bind:               bindURL,
		ControllerEndpoint: controllerEndpoint,
		LeaseEndpoint:      l.leaseAddress,
		Key:                model.NewLocalRunnerKey(keySuffix),
		Deployment:         deploymentKey,
		DebugPort:          debugPort,
		DevEndpoint:        devURI,
		DevRunnerInfoFile:  devRunnerInfoFile,
	}

	simpleName := fmt.Sprintf("runner%d", keySuffix)
	if err := kong.ApplyDefaults(&config, kong.Vars{
		"deploymentdir": filepath.Join(l.cacheDir, "ftl-runner", simpleName, "deployments"),
		// TODO: This doesn't seem like it should be here.
		"language": "go,kotlin,java",
	}); err != nil {
		return fmt.Errorf("failed to apply defaults: %w", err)
	}
	config.HeartbeatPeriod = time.Second
	config.HeartbeatJitter = time.Millisecond * 100

	runnerCtx := log.ContextWithLogger(ctx, logger.Scope(simpleName).Module(info.module))

	runnerCtx, cancel := context.WithCancel(runnerCtx)
	info.runner = optional.Some(runnerInfo{cancelFunc: cancel, port: bind.Port, host: "127.0.0.1"})

	go func() {
		err := runner.Start(runnerCtx, config, l.storage)
		l.lock.Lock()
		defer l.lock.Unlock()
		if devEndpoint != nil {
			// Runner is complete, clear the deployment key
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
