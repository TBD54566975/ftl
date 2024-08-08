package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	osExec "os/exec" //nolint:depguard
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/scaling/localscaling"
	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type serveCmd struct {
	Bind                *url.URL             `help:"Starting endpoint to bind to and advertise to. Each controller, ingress and runner will increment the port by 1" default:"http://127.0.0.1:8891"`
	DBPort              int                  `help:"Port to use for the database." default:"15432"`
	Recreate            bool                 `help:"Recreate the database even if it already exists." default:"false"`
	Controllers         int                  `short:"c" help:"Number of controllers to start." default:"1"`
	Background          bool                 `help:"Run in the background." default:"false"`
	Stop                bool                 `help:"Stop the running FTL instance. Can be used with --background to restart the server" default:"false"`
	StartupTimeout      time.Duration        `help:"Timeout for the server to start up." default:"1m"`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	DatabaseImage       string               `help:"The container image to start for the database" default:"postgres:15.4" env:"FTL_DATABASE_IMAGE" hidden:""`
	controller.CommonConfig
}

const ftlContainerName = "ftl-db-1"
const ftlRunningErrorMsg = "FTL is already running. Use 'ftl serve --stop' to stop it"

func (s *serveCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	return s.run(ctx, projConfig, optional.None[chan bool]())
}

func (s *serveCmd) run(ctx context.Context, projConfig projectconfig.Config, initialised optional.Option[chan bool]) error {
	logger := log.FromContext(ctx)
	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)

	if s.Background {
		if s.Stop {
			// allow usage of --background and --stop together to "restart" the background process
			_ = KillBackgroundServe(logger) //nolint:errcheck // ignore error here if the process is not running
		}

		if err := runInBackground(logger); err != nil {
			return err
		}

		err := waitForControllerOnline(ctx, s.StartupTimeout, client)
		if err != nil {
			return err
		}

		os.Exit(0)
	}

	if s.Stop {
		return KillBackgroundServe(logger)
	}

	if s.isRunning(ctx, client) {
		return errors.New(ftlRunningErrorMsg)
	}

	logger.Infof("Starting FTL with %d controller(s)", s.Controllers)

	// Bring up the DB and DAL.
	dsn, err := s.setupDB(ctx, s.DatabaseImage)
	if err != nil {
		return err
	}

	err = observability.Init(ctx, false, "", "ftl-serve", ftl.Version, s.ObservabilityConfig)
	if err != nil {
		return fmt.Errorf("observability init failed: %w", err)
	}

	wg, ctx := errgroup.WithContext(ctx)

	bindAllocator, err := bind.NewBindAllocator(s.Bind)
	if err != nil {
		return err
	}

	controllerAddresses := make([]*url.URL, 0, s.Controllers)
	ingressAddresses := make([]*url.URL, 0, s.Controllers)
	for range s.Controllers {
		ingressAddresses = append(ingressAddresses, bindAllocator.Next())
		controllerAddresses = append(controllerAddresses, bindAllocator.Next())
	}

	runnerScaling, err := localscaling.NewLocalScaling(bindAllocator, controllerAddresses)
	if err != nil {
		return err
	}
	for i := range s.Controllers {
		config := controller.Config{
			CommonConfig: s.CommonConfig,
			Bind:         controllerAddresses[i],
			IngressBind:  ingressAddresses[i],
			Key:          model.NewLocalControllerKey(i),
			DSN:          dsn,
		}
		config.SetDefaults()
		config.ModuleUpdateFrequency = time.Second * 1

		scope := fmt.Sprintf("controller%d", i)
		controllerCtx := log.ContextWithLogger(ctx, logger.Scope(scope))

		// create config manager for controller
		cr := cf.ProjectConfigResolver[cf.Configuration]{Config: projConfig.Path}
		cm, err := cf.NewConfigurationManager(controllerCtx, cr)
		if err != nil {
			return fmt.Errorf("could not create config manager: %w", err)
		}
		controllerCtx = cf.ContextWithConfig(controllerCtx, cm)

		// create secrets manager for controller
		sr := cf.ProjectConfigResolver[cf.Secrets]{Config: projConfig.Path}
		sm, err := cf.NewSecretsManager(controllerCtx, sr, cli.Vault, projConfig.Path)
		if err != nil {
			return fmt.Errorf("could not create secrets manager: %w", err)
		}
		controllerCtx = cf.ContextWithSecrets(controllerCtx, sm)

		// Bring up the DB connection and DAL.
		pool, err := pgxpool.New(ctx, config.DSN)
		if err != nil {
			return fmt.Errorf("failed to bring up DB connection: %w", err)
		}
		encryptors, err := config.EncryptionKeys.Encryptors(false)
		if err != nil {
			return fmt.Errorf("failed to create encryptors: %w", err)
		}

		wg.Go(func() error {
			if err := controller.Start(controllerCtx, config, runnerScaling, pool, encryptors); err != nil {
				logger.Errorf(err, "controller%d failed: %v", i, err)
				return fmt.Errorf("controller%d failed: %w", i, err)
			}
			return nil
		})
	}

	// Wait for controller to start, then run startup commands.
	start := time.Now()
	if err := waitForControllerOnline(ctx, s.StartupTimeout, client); err != nil {
		return fmt.Errorf("controller failed to start: %w", err)
	}
	logger.Infof("Controller started in %s", time.Since(start))

	if len(projConfig.Commands.Startup) > 0 {
		for _, cmd := range projConfig.Commands.Startup {
			logger.Debugf("Executing startup command: %s", cmd)
			if err := exec.Command(ctx, log.Info, ".", "bash", "-c", cmd).Run(); err != nil {
				return fmt.Errorf("startup command failed: %w", err)
			}
		}
	}

	if ch, ok := initialised.Get(); ok {
		ch <- true
	}

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("serve failed: %w", err)
	}

	return nil
}

func runInBackground(logger *log.Logger) error {
	if running, err := isServeRunning(logger); err != nil {
		return fmt.Errorf("failed to check if FTL is running: %w", err)
	} else if running {
		logger.Warnf(ftlRunningErrorMsg)
		return nil
	}

	args := make([]string, 0, len(os.Args))
	for _, arg := range os.Args[1:] {
		if arg == "--background" || arg == "--stop" {
			continue
		}
		args = append(args, arg)
	}

	cmd := osExec.Command(os.Args[0], args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background process: %w", err)
	}

	pidFilePath, err := pidFilePath()
	if err != nil {
		return fmt.Errorf("failed to get pid file path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(pidFilePath), 0750); err != nil {
		return fmt.Errorf("failed to create directory for pid file: %w", err)
	}

	if err := os.WriteFile(pidFilePath, []byte(strconv.Itoa(cmd.Process.Pid)), 0600); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}

	logger.Infof("`ftl serve` running in background with pid: %d", cmd.Process.Pid)
	return nil
}

func KillBackgroundServe(logger *log.Logger) error {
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

	if err := os.Remove(pidFilePath); err != nil {
		logger.Errorf(err, "Failed to remove pid file: %v", err)
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		if !errors.Is(err, syscall.ESRCH) {
			return err
		}
	}

	logger.Infof("`ftl serve` stopped (pid: %d)", pid)
	return nil
}

func pidFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".ftl", "ftl-serve.pid"), nil
}

func getPIDFromPath(path string) (int, error) {
	pidBytes, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func isServeRunning(logger *log.Logger) (bool, error) {
	pidFilePath, err := pidFilePath()
	if err != nil {
		return false, err
	}

	pid, err := getPIDFromPath(pidFilePath)
	if err != nil || pid == 0 {
		return false, err
	}

	err = syscall.Kill(pid, 0)
	if err != nil {
		if errors.Is(err, syscall.ESRCH) {
			logger.Infof("Process with PID %d does not exist.", pid)
			return false, nil
		}
		if errors.Is(err, syscall.EPERM) {
			logger.Infof("Process with PID %d exists but no permission to signal it.", pid)
			return true, nil
		}
		return false, err
	}

	return true, nil
}

func (s *serveCmd) setupDB(ctx context.Context, image string) (string, error) {
	logger := log.FromContext(ctx)

	recreate := s.Recreate
	port := s.DBPort

	exists, err := container.DoesExist(ctx, ftlContainerName, optional.Some(image))
	if err != nil {
		return "", err
	}

	if !exists {
		logger.Debugf("Creating docker container '%s' for postgres db", ftlContainerName)

		// check if port s.DBPort is already in use
		if l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.DBPort)); err != nil {
			return "", fmt.Errorf("port %d is already in use", s.DBPort)
		} else if err = l.Close(); err != nil {
			return "", fmt.Errorf("failed to close listener: %w", err)
		}

		err = container.RunDB(ctx, ftlContainerName, s.DBPort, image)
		if err != nil {
			return "", err
		}

		recreate = true
	} else {
		// Start the existing container
		err = container.Start(ctx, ftlContainerName)
		if err != nil {
			return "", err
		}

		// Grab the port from the existing container
		port, err = container.GetContainerPort(ctx, ftlContainerName, 5432)
		if err != nil {
			return "", err
		}

		logger.Debugf("Reusing existing docker container %s on port %d for postgres db", ftlContainerName, port)
	}

	err = container.PollContainerHealth(ctx, ftlContainerName, 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("db container failed to be healthy: %w", err)
	}

	dsn := fmt.Sprintf("postgres://postgres:secret@localhost:%d/ftl?sslmode=disable", port)
	logger.Debugf("Postgres DSN: %s", dsn)

	_, err = databasetesting.CreateForDevel(ctx, dsn, recreate)
	if err != nil {
		return "", err
	}

	return dsn, nil
}

// waitForControllerOnline polls the controller service until it is online.
func waitForControllerOnline(ctx context.Context, startupTimeout time.Duration, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Waiting %s for controller to be online", startupTimeout)

	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{}))
			if err != nil {
				logger.Tracef("Error getting status, retrying...: %v", err)
				continue // retry
			}

			return nil

		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logger.Errorf(ctx.Err(), "Timeout reached while polling for controller status")
			}
			return ctx.Err()
		}
	}
}

func (s *serveCmd) isRunning(ctx context.Context, client ftlv1connect.ControllerServiceClient) bool {
	_, err := client.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	return err == nil
}
