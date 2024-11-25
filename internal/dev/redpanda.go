package dev

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"sync"
	"time"

	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

const redPandaContainerName = "ftl-redpanda-1"

//go:embed docker-compose.redpanda.yml
var dockerCompose []byte
var dockerComposeLock sync.Mutex

func SetUpRedPanda(ctx context.Context) error {
	// A lock is used to provent Docker compose getting confused, which happens when we bring redpanda up
	// multiple times simultaneously.
	dockerComposeLock.Lock()
	defer dockerComposeLock.Unlock()

	cmd := exec.Command(ctx, log.Debug, ".", "docker", "compose", "-f", "-", "-p", "ftl", "up", "-d", "--wait")
	cmd.Stdin = bytes.NewReader(dockerCompose)
	if err := cmd.RunStderrError(ctx); err != nil {
		return fmt.Errorf("failed to run docker compose up: %w", err)
	}

	err := container.PollContainerHealth(ctx, redPandaContainerName, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("redpanda container failed to be healthy: %w", err)
	}
	return nil
}
