package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/runner"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"golang.org/x/sync/errgroup"
)

type serveCmd struct {
	Bind        *url.URL `help:"Starting endpoint to bind to and advertise to. Each controller and runner will increment the port by 1" default:"http://localhost:8892"`
	Controllers int      `short:"c" help:"Number of controllers to start." default:"1"`
	Runners     int      `short:"r" help:"Number of runners to start." default:"10"`
}

func (s *serveCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Infof("Starting %d controller(s) and %d runner(s)", s.Controllers, s.Runners)

	wg, ctx := errgroup.WithContext(ctx)

	controllerAddresses := make([]*url.URL, 0, s.Controllers)
	nextBind := s.Bind

	for i := 0; i < s.Controllers; i++ {
		controllerAddresses = append(controllerAddresses, nextBind)
		config := controller.Config{
			Bind: nextBind,
		}
		if err := kong.ApplyDefaults(&config); err != nil {
			return errors.WithStack(err)
		}

		scope := fmt.Sprintf("controller-%d", i)
		controllerCtx := log.ContextWithLogger(ctx, logger.Scope(scope))

		wg.Go(func() error { return controller.Start(controllerCtx, config) })

		var err error
		nextBind, err = incrementPort(nextBind)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return errors.WithStack(err)
	}

	for i := 0; i < s.Runners; i++ {
		controllerEndpoint := controllerAddresses[i%len(controllerAddresses)]
		fmt.Printf("controllerEndpoint: %s runner: %s\n", controllerEndpoint, nextBind)
		config := runner.Config{
			Bind:               nextBind,
			ControllerEndpoint: controllerEndpoint,
		}

		name := fmt.Sprintf("runner%d", i)
		if err := kong.ApplyDefaults(&config, kong.Vars{
			"deploymentdir": filepath.Join(cacheDir, "ftl-runner", name, "deployments"),
			"language":      "go,kotlin",
		}); err != nil {
			return errors.WithStack(err)
		}

		// Create a readable ULID for the runner.
		var ulid [16]byte
		binary.BigEndian.PutUint32(ulid[10:], uint32(i))
		ulidStr := fmt.Sprintf("%025X", ulid)
		err := config.Key.Scan(ulidStr)
		if err != nil {
			return errors.WithStack(err)
		}

		runnerCtx := log.ContextWithLogger(ctx, logger.Scope(name))

		wg.Go(func() error { return runner.Start(runnerCtx, config) })

		nextBind, err = incrementPort(nextBind)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if err := wg.Wait(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func incrementPort(baseURL *url.URL) (*url.URL, error) {
	newURL := *baseURL

	newPort, err := strconv.Atoi(newURL.Port())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	newURL.Host = fmt.Sprintf("%s:%d", baseURL.Hostname(), newPort+1)
	return &newURL, nil
}
