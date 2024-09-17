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

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/runner"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

var _ scaling.RunnerScaling = &localScaling{}

type localScaling struct {
	lock     sync.Mutex
	cacheDir string
	// Module -> Deployments -> info
	runners map[string]map[string]*deploymentInfo

	portAllocator       *bind.BindAllocator
	controllerAddresses []*url.URL

	prevRunnerSuffix int
}

func (l *localScaling) Start(ctx context.Context, endpoint url.URL, leaser leases.Leaser) error {
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
	replicas int32
	key      string
}
type runnerInfo struct {
	cancelFunc context.CancelFunc
	port       string
}

func NewLocalScaling(portAllocator *bind.BindAllocator, controllerAddresses []*url.URL) (scaling.RunnerScaling, error) {

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	local := localScaling{
		lock:                sync.Mutex{},
		cacheDir:            cacheDir,
		runners:             map[string]map[string]*deploymentInfo{},
		portAllocator:       portAllocator,
		controllerAddresses: controllerAddresses,
		prevRunnerSuffix:    -1,
	}

	return &local, nil
}

func (l *localScaling) handleSchemaChange(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
	if msg.DeploymentKey == "" {
		// Builtins don't have deployments
		return nil
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	logger := log.FromContext(ctx).Scope("localScaling")
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Infof("Handling schema change for %s", msg.DeploymentKey)
	moduleDeployments := l.runners[msg.ModuleName]
	if moduleDeployments == nil {
		moduleDeployments = map[string]*deploymentInfo{}
		l.runners[msg.ModuleName] = moduleDeployments
	}
	deploymentRunners := moduleDeployments[msg.DeploymentKey]
	if deploymentRunners == nil {
		deploymentRunners = &deploymentInfo{runner: optional.None[runnerInfo](), key: msg.DeploymentKey}
		moduleDeployments[msg.DeploymentKey] = deploymentRunners
	}

	switch msg.ChangeType {
	case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:
		deploymentRunners.replicas = msg.Schema.Runtime.MinReplicas
	case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
		deploymentRunners.replicas = 0
		delete(moduleDeployments, msg.DeploymentKey)
	}
	return l.reconcileRunners(ctx, deploymentRunners)
}

func (l *localScaling) reconcileRunners(ctx context.Context, deploymentRunners *deploymentInfo) error {
	// Must be called under lock
	logger := log.FromContext(ctx)
	if deploymentRunners.replicas > 0 && !deploymentRunners.runner.Ok() {
		if err := l.startRunner(ctx, deploymentRunners.key, deploymentRunners); err != nil {
			logger.Errorf(err, "Failed to start runner")
			return err
		}
	} else if deploymentRunners.replicas == 0 && deploymentRunners.runner.Ok() {
		deploymentRunners.runner.MustGet().cancelFunc()
		deploymentRunners.runner = optional.None[runnerInfo]()
	}
	return nil
}

func (l *localScaling) startRunner(ctx context.Context, deploymentKey string, info *deploymentInfo) error {
	controllerEndpoint := l.controllerAddresses[len(l.runners)%len(l.controllerAddresses)]

	bind := l.portAllocator.Next()
	keySuffix := l.prevRunnerSuffix + 1
	l.prevRunnerSuffix = keySuffix

	config := runner.Config{
		Bind:               bind,
		ControllerEndpoint: controllerEndpoint,
		Key:                model.NewLocalRunnerKey(keySuffix),
		Deployment:         deploymentKey,
	}

	simpleName := fmt.Sprintf("runner%d", keySuffix)
	if err := kong.ApplyDefaults(&config, kong.Vars{
		"deploymentdir": filepath.Join(l.cacheDir, "ftl-runner", simpleName, "deployments"),
		"language":      "go,kotlin,rust,java",
	}); err != nil {
		return fmt.Errorf("failed to apply defaults: %w", err)
	}
	config.HeartbeatPeriod = time.Second
	config.HeartbeatJitter = time.Millisecond * 100

	logger := log.FromContext(ctx)
	runnerCtx := log.ContextWithLogger(ctx, logger.Scope(simpleName))

	runnerCtx, cancel := context.WithCancel(runnerCtx)
	info.runner = optional.Some(runnerInfo{cancelFunc: cancel, port: bind.Port()})

	go func() {
		logger.Debugf("Starting runner: %s", config.Key)
		err := runner.Start(runnerCtx, config)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf(err, "Runner failed: %s", err)
		}
		l.lock.Lock()
		defer l.lock.Unlock()
		info.runner = optional.None[runnerInfo]()
		err = l.reconcileRunners(ctx, info)
		if err != nil {
			logger.Errorf(err, "Failed to reconcile runners")
		}
	}()
	return nil
}
