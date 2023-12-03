package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/common/bind"
	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/scaling/localscaling"
	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
)

type serveCmd struct {
	Bind         *url.URL   `help:"Starting endpoint to bind to and advertise to. Each controller and runner will increment the port by 1" default:"http://localhost:8892"`
	AllowOrigins []*url.URL `help:"Allow CORS requests to ingress endpoints from these origins." env:"FTL_CONTROLLER_ALLOW_ORIGIN"`
	DBPort       int        `help:"Port to use for the database." default:"54320"`
	Recreate     bool       `help:"Recreate the database even if it already exists." default:"false"`
	Controllers  int        `short:"c" help:"Number of controllers to start." default:"1"`
	Runners      int        `short:"r" help:"Number of runners to start." default:"0"`
}

const ftlContainerName = "ftl-db-1"

func (s *serveCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)

	dsn, err := s.setupDB(ctx)
	if err != nil {
		return err
	}

	logger.Infof("Starting %d controller(s) and %d runner(s)", s.Controllers, s.Runners)

	wg, ctx := errgroup.WithContext(ctx)

	bindAllocator, err := bind.NewBindAllocator(s.Bind)
	if err != nil {
		return err
	}

	controllerAddresses := make([]*url.URL, 0, s.Controllers)
	for i := 0; i < s.Controllers; i++ {
		controllerAddresses = append(controllerAddresses, bindAllocator.Next())
	}

	runnerScaling, err := localscaling.NewLocalScaling(bindAllocator, controllerAddresses)
	if err != nil {
		return err
	}

	for i := 0; i < s.Controllers; i++ {
		i := i
		config := controller.Config{
			Bind:         controllerAddresses[i],
			DSN:          dsn,
			AllowOrigins: s.AllowOrigins,
		}
		if err := kong.ApplyDefaults(&config); err != nil {
			return err
		}

		scope := fmt.Sprintf("controller%d", i)
		controllerCtx := log.ContextWithLogger(ctx, logger.Scope(scope))

		wg.Go(func() error {
			if err := controller.Start(controllerCtx, config, runnerScaling); err != nil {
				return fmt.Errorf("controller%d failed: %w", i, err)
			}
			return nil
		})
	}

	err = runnerScaling.SetReplicas(ctx, s.Runners, nil)
	if err != nil {
		return err
	}

	if err := wg.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *serveCmd) setupDB(ctx context.Context) (string, error) {
	logger := log.FromContext(ctx)

	nameFlag := fmt.Sprintf("name=^/%s$", ftlContainerName)
	output, err := exec.Capture(ctx, ".", "docker", "ps", "-a", "--filter", nameFlag, "--format", "{{.Names}}")
	if err != nil {
		logger.Errorf(err, "%s", output)
		return "", err
	}

	recreate := s.Recreate
	port := fmt.Sprintf("%d", s.DBPort)

	if len(output) == 0 {
		logger.Infof("Creating docker container '%s' for postgres db", ftlContainerName)

		// check if port s.DBPort is already in use
		_, err := exec.Capture(ctx, ".", "sh", "-c", fmt.Sprintf("lsof -i:%d", s.DBPort))
		if err == nil {
			return "", fmt.Errorf("port %d is already in use", s.DBPort)
		}

		err = exec.Command(ctx, logger.GetLevel(), "./", "docker", "run",
			"-d", // run detached so we can follow with other commands
			"--name", ftlContainerName,
			"--user", "postgres",
			"--restart", "always",
			"-e", "POSTGRES_PASSWORD=secret",
			"-p", fmt.Sprintf("%s:5432", port),
			"--health-cmd=pg_isready",
			"--health-interval=1s",
			"--health-timeout=60s",
			"--health-retries=60",
			"--health-start-period=80s",
			"postgres:latest", "postgres",
		).Run()

		if err != nil {
			return "", err
		}

		err = pollContainerHealth(ctx, ftlContainerName, 10*time.Second)
		if err != nil {
			return "", err
		}

		recreate = true
	} else {
		// Start the existing container
		_, err = exec.Capture(ctx, ".", "docker", "start", ftlContainerName)
		if err != nil {
			return "", err
		}

		// Grab the port from the existing container
		portOutput, err := exec.Capture(ctx, ".", "docker", "port", ftlContainerName, "5432/tcp")
		if err != nil {
			logger.Errorf(err, "%s", portOutput)
			return "", err
		}
		port = slices.Reduce(strings.Split(string(portOutput), "\n"), "", func(port string, line string) string {
			if parts := strings.Split(line, ":"); len(parts) == 2 {
				return parts[1]
			}
			return port
		})

		logger.Infof("Reusing existing docker container %q on port %q for postgres db", ftlContainerName, port)
	}

	dsn := fmt.Sprintf("postgres://postgres:secret@localhost:%s/ftl?sslmode=disable", port)
	logger.Infof("Postgres DSN: %s", dsn)

	_, err = databasetesting.CreateForDevel(ctx, dsn, recreate)
	if err != nil {
		return "", err
	}

	return dsn, nil
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
				return err
			}

			status := strings.TrimSpace(string(output))
			if status == "healthy" {
				return nil
			}
		}
	}
}
