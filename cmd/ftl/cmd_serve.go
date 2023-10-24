package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/runner"
)

type serveCmd struct {
	Bind        *url.URL `help:"Starting endpoint to bind to and advertise to. Each controller and runner will increment the port by 1" default:"http://localhost:8892"`
	Controllers int      `short:"c" help:"Number of controllers to start." default:"1"`
	Runners     int      `short:"r" help:"Number of runners to start." default:"10"`
}

const ftlContainerName = "ftl-db"

func (s *serveCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)

	dsn, err := setupDB(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	logger.Infof("Starting %d controller(s) and %d runner(s)", s.Controllers, s.Runners)

	wg, ctx := errgroup.WithContext(ctx)

	controllerAddresses := make([]*url.URL, 0, s.Controllers)
	nextBind := s.Bind

	for i := 0; i < s.Controllers; i++ {
		i := i
		controllerAddresses = append(controllerAddresses, nextBind)
		config := controller.Config{
			Bind: nextBind,
			DSN:  dsn,
		}
		if err := kong.ApplyDefaults(&config); err != nil {
			return errors.WithStack(err)
		}

		scope := fmt.Sprintf("controller%d", i)
		controllerCtx := log.ContextWithLogger(ctx, logger.Scope(scope))

		wg.Go(func() error {
			return errors.Wrapf(controller.Start(controllerCtx, config), "controller%d failed", i)
		})

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
		i := i
		controllerEndpoint := controllerAddresses[i%len(controllerAddresses)]
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

		wg.Go(func() error {
			return errors.Wrapf(runner.Start(runnerCtx, config), "runner%d failed", i)
		})

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

func setupDB(ctx context.Context) (string, error) {
	logger := log.FromContext(ctx)
	logger.Infof("Checking for FTL database")

	nameFlag := fmt.Sprintf("name=^/%s$", ftlContainerName)
	output, err := exec.Capture(ctx, ".", "docker", "ps", "-a", "--filter", nameFlag, "--format", "{{.Names}}")
	if err != nil {
		logger.Errorf(err, "%s", output)
		return "", errors.WithStack(err)
	}

	recreate := false

	if len(output) == 0 {
		logger.Infof("Creating FTL database")

		err = exec.Command(ctx, logger.GetLevel(), "./", "docker", "run",
			"-d", // run detached so we can follow with other commands
			"--name", ftlContainerName,
			"--user", "postgres",
			"--restart", "always",
			"-e", "POSTGRES_PASSWORD=secret",
			"-p", "5432", // dynamically allocate port
			"--health-cmd=pg_isready",
			"--health-interval=1s",
			"--health-timeout=60s",
			"--health-retries=60",
			"--health-start-period=80s",
			"postgres:latest", "postgres",
		).Run()

		if err != nil {
			return "", errors.WithStack(err)
		}

		err = pollContainerHealth(ctx, ftlContainerName, 10*time.Second)
		if err != nil {
			return "", err
		}

		recreate = true
	}

	// grab the port from docker for this container
	port, err := exec.Capture(ctx, ".", "docker", "inspect", "--format", "{{ (index (index .NetworkSettings.Ports \"5432/tcp\") 0).HostPort }}", ftlContainerName)
	if err != nil {
		return "", errors.WithStack(err)
	}

	dsn := fmt.Sprintf("postgres://postgres:secret@localhost:%s/%s?sslmode=disable", strings.TrimSpace(string(port)), ftlContainerName)
	dsnFlag := fmt.Sprintf("--dsn=%s", dsn)

	if recreate {
		logger.Infof("Initializing FTL schema")
		err = exec.Command(ctx, logger.GetLevel(), ".", "ftl-initdb", "--recreate", dsnFlag).Run()
	} else {
		err = exec.Command(ctx, logger.GetLevel(), ".", "ftl-initdb", dsnFlag).Run()
	}
	if err != nil {
		return "", errors.WithStack(err)
	}

	return dsn, nil
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

func pollContainerHealth(ctx context.Context, containerName string, timeout time.Duration) error {
	logger := log.FromContext(ctx)
	logger.Infof("Waiting for %s to be healthy", containerName)

	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-pollCtx.Done():
			return errors.New("timed out waiting for container to be healthy")
		case <-time.After(1 * time.Millisecond):
			output, err := exec.Capture(pollCtx, ".", "docker", "inspect", "--format", "{{.State.Health.Status}}", containerName)
			if err != nil {
				return errors.WithStack(err)
			}

			status := strings.TrimSpace(string(output))
			if status == "healthy" {
				return nil
			}
		}
	}
}
