package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	osExec "os/exec" //nolint:depguard
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/scaling/localscaling"
	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

type serveCmd struct {
	Bind         *url.URL   `help:"Starting endpoint to bind to and advertise to. Each controller and runner will increment the port by 1" default:"http://localhost:8892"`
	AllowOrigins []*url.URL `help:"Allow CORS requests to ingress endpoints from these origins." env:"FTL_CONTROLLER_ALLOW_ORIGIN"`
	DBPort       int        `help:"Port to use for the database." default:"54320"`
	Recreate     bool       `help:"Recreate the database even if it already exists." default:"false"`
	Controllers  int        `short:"c" help:"Number of controllers to start." default:"1"`
	Runners      int        `short:"r" help:"Number of runners to start." default:"0"`
	Background   bool       `help:"Run in the background." default:"false"`
	Stop         bool       `help:"Stop the running FTL instance." default:"false"`
}

const ftlContainerName = "ftl-db-1"

func (s *serveCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)

	if s.Background {
		runInBackground(logger)
		os.Exit(0)
	}

	if s.Stop {
		return killBackgroundProcess(logger)
	}

	logger.Infof("Starting FTL with %d controller(s) and %d runner(s)", s.Controllers, s.Runners)

	dsn, err := s.setupDB(ctx)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("serve failed: %w", err)
	}
	return nil
}

func runInBackground(logger *log.Logger) {
	pidFilePath, err := pidFilePath()
	if err != nil {
		logger.Errorf(err, "failed to get pid file path")
		return
	}

	if existingPID, err := getPIDFromPath(pidFilePath); err == nil && existingPID != 0 {
		logger.Warnf("'ftl serve' is already running in the background. Use --stop to stop it.")
		return
	}

	args := make([]string, 0, len(os.Args))
	for _, arg := range os.Args[1:] {
		if arg == "--background" {
			continue
		}
		args = append(args, arg)
	}

	cmd := osExec.Command(os.Args[0], args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		logger.Errorf(err, "failed to start background process")
	}

	if err := os.MkdirAll(filepath.Dir(pidFilePath), 0750); err != nil {
		logger.Errorf(err, "failed to create directory for pid file")
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFilePath, []byte(fmt.Sprintf("%d", pid)), 0600); err != nil {
		logger.Errorf(err, "failed to write pid file")
	}

	logger.Infof("FTL running in background with pid: %d\n", pid)
}

func killBackgroundProcess(logger *log.Logger) error {
	pidFilePath, err := pidFilePath()
	if err != nil {
		logger.Infof("No background process found")
		return err
	}

	pid, err := getPIDFromPath(pidFilePath)
	if err != nil || pid == 0 {
		logger.Debugf("FTL serve is not running in the background")
		return nil
	}

	err = os.Remove(pidFilePath)
	if err != nil {
		logger.Errorf(err, "failed to remove pid file")
	}

	return syscall.Kill(pid, syscall.SIGTERM)
}

func getPIDFromPath(path string) (int, error) {
	pid, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	pidInt, err := strconv.Atoi(string(pid))
	if err != nil {
		return 0, err
	}

	return pidInt, nil
}

func pidFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".ftl", "ftl-serve.pid"), nil
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
	port := strconv.Itoa(s.DBPort)

	if len(output) == 0 {
		logger.Debugf("Creating docker container '%s' for postgres db", ftlContainerName)

		// check if port s.DBPort is already in use
		_, err := exec.Capture(ctx, ".", "sh", "-c", fmt.Sprintf("lsof -i:%d", s.DBPort))
		if err == nil {
			return "", fmt.Errorf("port %d is already in use", s.DBPort)
		}

		err = exec.Command(ctx, log.Debug, "./", "docker", "run",
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
		).RunBuffered(ctx)
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

		logger.Debugf("Reusing existing docker container %q on port %q for postgres db", ftlContainerName, port)
	}

	err = pollContainerHealth(ctx, ftlContainerName, 10*time.Second)
	if err != nil {
		return "", err
	}

	dsn := fmt.Sprintf("postgres://postgres:secret@localhost:%s/ftl?sslmode=disable", port)
	logger.Debugf("Postgres DSN: %s", dsn)

	_, err = databasetesting.CreateForDevel(ctx, dsn, recreate)
	if err != nil {
		return "", err
	}

	return dsn, nil
}

func pollContainerHealth(ctx context.Context, containerName string, timeout time.Duration) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Waiting for %s to be healthy", containerName)

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
