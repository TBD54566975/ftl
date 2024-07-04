//go:build integration

package main

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	. "github.com/TBD54566975/ftl/integration"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestBox(t *testing.T) {
	t.Skip("skipping due to timeouts")

	// Need a longer timeout to wait for FTL inside Docker.
	t.Setenv("FTL_INTEGRATION_TEST_TIMEOUT", "30s")
	Infof("Building local ftl0/ftl-box:latest Docker image")
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	err := exec.Command(ctx, log.Debug, "../..", "docker", "build", "-t", "ftl0/ftl-box:latest", "--progress=plain", "--platform=linux/amd64", "-f", "Dockerfile.box", ".").Run()
	assert.NoError(t, err)
	RunWithoutController(t, "",
		CopyModule("time"),
		CopyModule("echo"),
		Exec("ftl", "box", "echo", "--compose=echo-compose.yml"),
		Exec("docker", "compose", "-f", "echo-compose.yml", "up", "--wait"),
		Call("echo", "echo", Obj{"name": "Alice"}, nil),
		Exec("docker", "compose", "-f", "echo-compose.yml", "down", "--rmi", "local"),
	)
}
