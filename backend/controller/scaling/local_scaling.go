package scaling

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/backend/common/bind"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/runner"
)

var _ RunnerScaling = (*LocalScaling)(nil)

type LocalScaling struct {
	lock     sync.Mutex
	cacheDir string
	runners  map[model.RunnerKey]context.CancelFunc

	portAllocator       *bind.BindAllocator
	controllerAddresses []*url.URL
}

func NewLocalScaling(portAllocator *bind.BindAllocator, controllerAddresses []*url.URL) (*LocalScaling, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &LocalScaling{
		lock:                sync.Mutex{},
		cacheDir:            cacheDir,
		runners:             map[model.RunnerKey]context.CancelFunc{},
		portAllocator:       portAllocator,
		controllerAddresses: controllerAddresses,
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
				return errors.WithStack(err)
			}
		}

		return nil
	}

	logger.Infof("Adding %d replicas", replicasToAdd)
	for i := 0; i < replicasToAdd; i++ {
		i := i

		controllerEndpoint := l.controllerAddresses[len(l.runners)%len(l.controllerAddresses)]
		config := runner.Config{
			Bind:               l.portAllocator.Next(),
			ControllerEndpoint: controllerEndpoint,
		}

		name := fmt.Sprintf("runner%d", i)
		if err := kong.ApplyDefaults(&config, kong.Vars{
			"deploymentdir": filepath.Join(l.cacheDir, "ftl-runner", name, "deployments"),
			"language":      "go,kotlin",
		}); err != nil {
			return errors.WithStack(err)
		}

		// Create a readable ULID for the runner.
		var ulid [16]byte
		binary.BigEndian.PutUint32(ulid[10:], uint32(len(l.runners)+1))
		ulidStr := fmt.Sprintf("%025X", ulid)
		err := config.Key.Scan(ulidStr)
		if err != nil {
			return errors.WithStack(err)
		}

		runnerCtx := log.ContextWithLogger(ctx, logger.Scope(name))

		runnerCtx, cancel := context.WithCancel(runnerCtx)
		l.runners[config.Key] = cancel

		go func() {
			logger.Infof("Starting runner: %s", config.Key)
			err := runner.Start(runnerCtx, config)
			if err != nil {
				logger.Errorf(err, "Error starting runner: %s", err)
			}
		}()
	}

	return nil
}

func (l *LocalScaling) remove(ctx context.Context, runner model.RunnerKey) error {
	log := log.FromContext(ctx)
	log.Infof("Removing runner: %s", runner)

	cancel, ok := l.runners[runner]
	if !ok {
		return errors.Errorf("runner %s not found", runner)
	}

	cancel()
	delete(l.runners, runner)

	return nil
}
