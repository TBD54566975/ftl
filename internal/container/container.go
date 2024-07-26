package container

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/types/once"
	"github.com/alecthomas/types/optional"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/TBD54566975/ftl/internal/log"
)

var dockerClient = once.Once(func(ctx context.Context) (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
})

func DoesExist(ctx context.Context, name string, image optional.Option[string]) (bool, error) {
	cli, err := dockerClient.Get(ctx)
	logger := log.FromContext(ctx)
	if err != nil {
		return false, err
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", name)),
	})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}
	if len(containers) == 0 {
		return false, nil
	}
	imageName, ok := image.Get()
	if !ok {
		return true, nil
	}
	for _, c := range containers {
		if c.Image != imageName {
			logger.Infof("possible database version mismatch, expecting to use container image %s for container with name %s, bit it was already running with image %s", image, name, c.Image)
			break
		}
	}
	return true, nil
}

// Pull pulls the given image.
func Pull(ctx context.Context, imageName string) error {
	cli, err := dockerClient.Get(ctx)
	if err != nil {
		return err
	}

	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull %s image: %w", imageName, err)
	}
	defer reader.Close()

	logger := log.FromContext(ctx)
	_, err = io.Copy(logger.WriterAt(log.Info), reader)
	if err != nil {
		return fmt.Errorf("failed to stream pull: %w", err)
	}

	return nil
}

// Run starts a new detached container with the given image, name, port map, and (optional) volume mount.
func Run(ctx context.Context, image, name string, hostPort, containerPort int, volume optional.Option[string]) error {
	cli, err := dockerClient.Get(ctx)
	if err != nil {
		return err
	}

	config := container.Config{
		Image: image,
	}

	containerNatPort := nat.Port(fmt.Sprintf("%d/tcp", containerPort))
	hostConfig := container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyAlways,
		},
		PortBindings: nat.PortMap{
			containerNatPort: []nat.PortBinding{
				{
					HostPort: strconv.Itoa(hostPort),
				},
			},
		},
	}
	if v, ok := volume.Get(); ok {
		hostConfig.Binds = []string{v}
	}

	created, err := cli.ContainerCreate(ctx, &config, &hostConfig, nil, nil, name)
	if err != nil {
		return fmt.Errorf("failed to create %s container: %w", name, err)
	}

	err = cli.ContainerStart(ctx, created.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start %s container: %w", name, err)
	}

	return nil
}

// RunDB runs a new detached postgres container with the given name and exposed port.
func RunDB(ctx context.Context, name string, port int, image string) error {
	cli, err := dockerClient.Get(ctx)
	if err != nil {
		return err
	}

	exists, err := DoesExist(ctx, name, optional.Some(image))
	if err != nil {
		return err
	}

	if !exists {
		err = Pull(ctx, image)
		if err != nil {
			return err
		}
	}

	config := container.Config{
		Image: image,
		Env:   []string{"POSTGRES_PASSWORD=secret"},
		User:  "postgres",
		Cmd:   []string{"postgres"},
		Healthcheck: &container.HealthConfig{
			Test:        []string{"CMD-SHELL", "pg_isready"},
			Interval:    time.Second,
			Retries:     60,
			Timeout:     60 * time.Second,
			StartPeriod: 80 * time.Second,
		},
	}

	hostConfig := container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyAlways,
		},
		PortBindings: nat.PortMap{
			"5432/tcp": []nat.PortBinding{
				{
					HostPort: strconv.Itoa(port),
				},
			},
		},
	}

	created, err := cli.ContainerCreate(ctx, &config, &hostConfig, nil, nil, name)
	if err != nil {
		return fmt.Errorf("failed to create db container: %w", err)
	}

	err = cli.ContainerStart(ctx, created.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start db container: %w", err)
	}

	return nil
}

// Start starts an existing container with the given name.
func Start(ctx context.Context, name string) error {
	cli, err := dockerClient.Get(ctx)
	if err != nil {
		return err
	}

	err = cli.ContainerStart(ctx, name, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

// Exec runs a command in the given container, stream to stderr. Return an error if the command fails.
func Exec(ctx context.Context, name string, command ...string) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Running command %q in container %q", command, name)

	cli, err := dockerClient.Get(ctx)
	if err != nil {
		return err
	}

	exec, err := cli.ContainerExecCreate(ctx, name, types.ExecConfig{
		Cmd:          command,
		AttachStderr: true,
		AttachStdout: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	attach, err := cli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("failed to attach exec: %w", err)
	}
	defer attach.Close()

	_, err = io.Copy(os.Stderr, attach.Reader)
	if err != nil {
		return fmt.Errorf("failed to stream exec: %w", err)
	}

	err = cli.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("failed to start exec: %w", err)
	}

	inspect, err := cli.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect exec: %w", err)
	}
	if inspect.ExitCode != 0 {
		return fmt.Errorf("exec failed with exit code %d", inspect.ExitCode)
	}

	return nil
}

// GetContainerPort returns the host TCP port of the given container's exposed port.
func GetContainerPort(ctx context.Context, name string, port int) (int, error) {
	cli, err := dockerClient.Get(ctx)
	if err != nil {
		return 0, err
	}

	inspect, err := cli.ContainerInspect(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect container: %w", err)
	}

	containerPort := fmt.Sprintf("%d/tcp", port)
	hostPort, ok := inspect.NetworkSettings.Ports[nat.Port(containerPort)]
	if !ok {
		return 0, fmt.Errorf("container port %q not found", containerPort)
	}

	if len(hostPort) == 0 {
		return 0, fmt.Errorf("container port %q not bound", containerPort)
	}

	return nat.Port(hostPort[0].HostPort).Int(), nil
}

// PollContainerHealth polls the given container until it is healthy or the timeout is reached.
func PollContainerHealth(ctx context.Context, containerName string, timeout time.Duration) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Waiting for %s to be healthy", containerName)

	cli, err := dockerClient.Get(ctx)
	if err != nil {
		return err
	}

	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-pollCtx.Done():
			return fmt.Errorf("timed out waiting for container to be healthy: %w", pollCtx.Err())

		case <-time.After(100 * time.Millisecond):
			inspect, err := cli.ContainerInspect(pollCtx, containerName)
			if err != nil {
				return fmt.Errorf("failed to inspect container: %w", err)
			}

			if inspect.State.Health.Status == types.Healthy {
				return nil
			}
		}
	}
}
