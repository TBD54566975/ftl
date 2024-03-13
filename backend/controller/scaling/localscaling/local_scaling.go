package localscaling

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/backend/controller/scaling"
	"github.com/TBD54566975/ftl/backend/runner"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

var _ scaling.RunnerScaling = (*LocalScaling)(nil)

type LocalScaling struct {
	lock     sync.Mutex
	cacheDir string
	runners  map[model.RunnerKey]context.CancelFunc

	portAllocator       *bind.BindAllocator
	controllerAddresses []*url.URL

	prevRunnerSuffix int
}

func NewLocalScaling(portAllocator *bind.BindAllocator, controllerAddresses []*url.URL) (*LocalScaling, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	return &LocalScaling{
		lock:                sync.Mutex{},
		cacheDir:            cacheDir,
		runners:             map[model.RunnerKey]context.CancelFunc{},
		portAllocator:       portAllocator,
		controllerAddresses: controllerAddresses,
		prevRunnerSuffix:    -1,
	}, nil
}

func (l *LocalScaling) SetReplicas(ctx context.Context, replicas int, idleRunners []model.RunnerKey) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	logger := log.FromContext(ctx)

	replicasToAdd := replicas - len(l.runners)

	if replicasToAdd <= 0 {
		replicasToRemove := -replicasToAdd

		for i := 0; i < replicasToRemove; i++ {
			if len(idleRunners) == 0 {
				return nil
			}
			runnerToRemove := idleRunners[len(idleRunners)-1]
			idleRunners = idleRunners[:len(idleRunners)-1]

			err := l.remove(ctx, runnerToRemove)
			if err != nil {
				return err
			}
		}

		return nil
	}

	logger.Debugf("Adding %d replicas", replicasToAdd)
	for i := 0; i < replicasToAdd; i++ {
		controllerEndpoint := l.controllerAddresses[len(l.runners)%len(l.controllerAddresses)]

		bind := l.portAllocator.Next()
		keySuffix := l.prevRunnerSuffix + 1
		l.prevRunnerSuffix = keySuffix

		config := runner.Config{
			Bind:               bind,
			ControllerEndpoint: controllerEndpoint,
			TemplateDir:        templateDir(ctx),
			Key:                model.NewLocalRunnerKey(keySuffix),
		}

		simpleName := fmt.Sprintf("runner%d", keySuffix)
		if err := kong.ApplyDefaults(&config, kong.Vars{
			"deploymentdir": filepath.Join(l.cacheDir, "ftl-runner", simpleName, "deployments"),
			"language":      "go,kotlin",
		}); err != nil {
			return err
		}

		runnerCtx := log.ContextWithLogger(ctx, logger.Scope(simpleName))

		runnerCtx, cancel := context.WithCancel(runnerCtx)
		l.runners[config.Key] = cancel

		go func() {
			logger.Debugf("Starting runner: %s", config.Key)
			err := runner.Start(runnerCtx, config)
			if err != nil && !errors.Is(err, context.Canceled) {
				logger.Errorf(err, "Runner failed: %s", err)
			}
		}()
	}

	return nil
}

func (l *LocalScaling) remove(ctx context.Context, runner model.RunnerKey) error {
	log := log.FromContext(ctx)
	log.Debugf("Removing runner: %s", runner)

	cancel, ok := l.runners[runner]
	if !ok {
		return fmt.Errorf("runner %s not found", runner)
	}

	cancel()
	delete(l.runners, runner)

	return nil
}
