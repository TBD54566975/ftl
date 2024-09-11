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

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/runner"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

type localScaling struct {
	lock     sync.Mutex
	cacheDir string
	// Module -> Deployments -> Runners -> Cancel Func
	runners map[string]map[string]*deploymentInfo

	portAllocator       *bind.BindAllocator
	controllerAddresses []*url.URL

	prevRunnerSuffix int
}

type deploymentInfo struct {
	runners  map[string]context.CancelFunc
	replicas int32
	key      string
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

	return func(ctx context.Context, endpoint url.URL, leaser leases.Leaser) error {
		scaling.BeginGrpcScaling(ctx, endpoint, leaser, local.handleSchemaChange)
		return nil
	}, nil
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
		deploymentRunners = &deploymentInfo{runners: map[string]context.CancelFunc{}, key: msg.DeploymentKey}
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
	existing := int32(len(deploymentRunners.runners))
	if existing < deploymentRunners.replicas {
		for i := existing; i < deploymentRunners.replicas; i++ {
			if err := l.startRunner(ctx, deploymentRunners.key, deploymentRunners); err != nil {
				logger.Errorf(err, "Failed to start runner")
				return err
			}
		}
	} else if existing > deploymentRunners.replicas {
		for _, cancelFunc := range deploymentRunners.runners {
			cancelFunc()
			existing--
		}
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
	info.runners[config.Key.String()] = cancel

	go func() {
		logger.Debugf("Starting runner: %s", config.Key)
		err := runner.Start(runnerCtx, config)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf(err, "Runner failed: %s", err)
		}
		l.lock.Lock()
		defer l.lock.Unlock()
		delete(info.runners, config.Key.String())
		err = l.reconcileRunners(ctx, info)
		if err != nil {
			logger.Errorf(err, "Failed to reconcile runners")
		}
	}()
	return nil
}
