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

func TestSecretImportExport(t *testing.T) {
	firstProjFile := "ftl-project.toml"
	secondProjFile := "ftl-project-2.toml"

	// use a pointer to keep track of the exported json so that i can be modified from within actions
	blank := ""
	exported := &blank

	RunWithoutController(t, "",
		// duplicate project file
		Exec("cp", firstProjFile, secondProjFile),
		// import into first project file
		Exec("ftl", "secret", "import", "--inline", "--config", firstProjFile, `
		{
			"test.one": 1,
			"test.two": "a string",
			"test2.three": {"key":"value"}
		}
		`),

		// export from first project file
		ExecWithOutput("ftl", []string{"secret", "export", "--config", firstProjFile}, func(output string) {
			*exported = output
			// make sure the exported json contains a secret (otherwise the test could pass with the first import doing nothing)
			assert.Contains(t, output, "test.one")
		}),

		// import into second project file
		// wrapped in a func to avoid capturing the initial valye of *exported
		func(t testing.TB, ic TestContext) {
			Exec("ftl", "secret", "import", *exported, "--inline", "--config", secondProjFile)(t, ic)
		},

		// export from second project file
		ExecWithOutput("ftl", []string{"secret", "export", "--config", secondProjFile}, func(output string) {
			// check that both exported the same json
			assert.Equal(t, *exported, output)
		}),
	)
}
