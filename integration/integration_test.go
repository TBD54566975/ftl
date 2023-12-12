//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"
	"time"

	"errors"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/rpc"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

const integrationTestTimeout = time.Second * 60

func TestIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	rootDir := filepath.Join(cwd, "..")

	modulesDir := filepath.Join(tmpDir, "modules")

	tests := []struct {
		name       string
		assertions assertions
	}{
		{name: "DeployTime", assertions: assertions{
			run("examples", "ftl-go", "deploy", "time"),
			deploymentExists("time"),
		}},
		{name: "CallTime", assertions: assertions{
			call("time", "time", obj{}, func(t testing.TB, resp obj) {
				assert.Equal(t, maps.Keys(resp), []string{"time"})
			}),
		}},
		{name: "DeployEchoKotlin", assertions: assertions{
			run(".", "ftl", "deploy", "examples/kotlin/ftl-module-echo"),
			deploymentExists("echo"),
		}},
		{name: "CallEchoKotlin", assertions: assertions{
			call("echo", "echo", obj{"name": "Alice"}, func(t testing.TB, resp obj) {
				message, ok := resp["message"].(string)
				assert.True(t, ok, "message is not a string")
				assert.True(t, regexp.MustCompile(`^Hello, Alice!`).MatchString(message), "%q does not match %q", message, `^Hello, Alice!`)
			}),
		}},
		{name: "InitNewKotlin", assertions: assertions{
			run(".", "ftl", "init", "kotlin", modulesDir, "echo2"),
			run(".", "ftl", "init", "kotlin", modulesDir, "echo3"),
		}},
		{name: "DeployNewKotlinEcho2", assertions: assertions{
			run(".", "ftl", "deploy", filepath.Join(modulesDir, "ftl-module-echo2")),
			deploymentExists("echo2"),
		}},
		{name: "CallEcho2", assertions: assertions{
			call("echo2", "echo", obj{"name": "Alice"}, func(t testing.TB, resp obj) {
				message, ok := resp["message"].(string)
				assert.True(t, ok, "message is not a string")
				assert.True(t, regexp.MustCompile(`^Hello, Alice!`).MatchString(message), "%q does not match %q", message, `^Hello, Alice!`)
			}),
		}},
		{name: "DeployNewKotlinEcho3", assertions: assertions{
			run(".", "ftl", "deploy", filepath.Join(modulesDir, "ftl-module-echo3")),
			deploymentExists("echo3"),
		}},
		{name: "CallEcho3", assertions: assertions{
			call("echo3", "echo", obj{"name": "Alice"}, func(t testing.TB, resp obj) {
				message, ok := resp["message"].(string)
				assert.True(t, ok, "message is not a string")
				assert.True(t, regexp.MustCompile(`^Hello, Alice!`).MatchString(message), "%q does not match %q", message, `^Hello, Alice!`)
			}),
		}},
		{name: "SchemaGenerateJS", assertions: assertions{
			run(".", "ftl", "schema", "generate", "integration/testdata/schema-generate", "build/schema-generate"),
			filesExist(file{"build/schema-generate/test.txt", "olleh"}),
		}},
	}

	// Build FTL binary
	logger := log.Configure(&logWriter{logger: t}, log.Config{Level: log.Debug})
	ctx := log.ContextWithLogger(context.Background(), logger)
	logger.Infof("Building ftl")
	binDir := filepath.Join(rootDir, "build", "release")
	err = exec.Command(ctx, log.Debug, rootDir, filepath.Join(rootDir, "bin", "bit"), "build/release/ftl", "**/*.jar").Run()
	assert.NoError(t, err)

	controller := rpc.Dial(ftlv1connect.NewControllerServiceClient, "http://localhost:8892", log.Debug)
	verbs := rpc.Dial(ftlv1connect.NewVerbServiceClient, "http://localhost:8892", log.Debug)

	ctx = startProcess(t, ctx, filepath.Join(binDir, "ftl"), "serve", "--recreate")

	ic := itContext{
		Context:    ctx,
		tmpDir:     tmpDir,
		rootDir:    rootDir,
		binDir:     binDir,
		controller: controller,
		verbs:      verbs,
	}

	ic.assertWithRetry(t, func(t testing.TB, ic itContext) error {
		_, err := ic.controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
		return err
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, assertion := range tt.assertions {
				ic.assertWithRetry(t, assertion)
			}
		})
	}
}

type assertion func(t testing.TB, ic itContext) error
type assertions []assertion

// Assertions

// Run a command in "dir" which is relative to the root directory of the project.
func run(dir, cmd string, args ...string) assertion {
	return func(t testing.TB, ic itContext) error {
		path := os.Getenv("PATH")
		path = ic.binDir + ":" + path
		cmd := exec.Command(ic, log.Debug, filepath.Join(ic.rootDir, dir), cmd, args...)
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env, "PATH="+path)
		return cmd.Run()
	}
}

func deploymentExists(name string) assertion {
	return status(func(t testing.TB, status *ftlv1.StatusResponse) {
		for _, deployment := range status.Deployments {
			if deployment.Schema.Name == "time" {
				return
			}
		}
		t.Fatal("time deployment not found")
	})
}

// Check status of the controller.
func status(check func(t testing.TB, status *ftlv1.StatusResponse)) assertion {
	return func(t testing.TB, ic itContext) error {
		status, err := ic.controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
		if err != nil {
			return err
		}
		check(t, status.Msg)
		return nil
	}
}

type file struct {
	path    string
	content string
}

// Assert that files exist in the temp dir.
func filesExist(files ...file) assertion {
	return func(t testing.TB, ic itContext) error {
		for _, file := range files {
			content, err := os.ReadFile(filepath.Join(ic.rootDir, file.path))
			if err != nil {
				return err
			}
			if string(content) != file.content {
				return fmt.Errorf("%s:\nExpected: %s\n  Actual: %s", file.path, file.content, string(content))
			}
		}
		return nil
	}
}

type obj map[string]any

func call[Resp any](module, verb string, req obj, onResponse func(t testing.TB, resp Resp)) assertion {
	return func(t testing.TB, ic itContext) error {
		jreq, err := json.Marshal(req)
		assert.NoError(t, err)

		cresp, err := ic.verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
			Verb: &schemapb.VerbRef{Module: module, Name: verb},
			Body: jreq,
		}))
		if err != nil {
			return err
		}

		if cresp.Msg.GetError() != nil {
			return errors.New(cresp.Msg.GetError().GetMessage())
		}

		var resp Resp
		err = json.Unmarshal(cresp.Msg.GetBody(), &resp)
		assert.NoError(t, err)

		onResponse(t, resp)
		return nil
	}
}

type itContext struct {
	context.Context
	tmpDir     string
	rootDir    string
	binDir     string // Where "ftl" binary is located.
	controller ftlv1connect.ControllerServiceClient
	verbs      ftlv1connect.VerbServiceClient
}

func (i itContext) assertWithRetry(t testing.TB, assertion assertion) {
	t.Helper()
	waitCtx, done := context.WithTimeout(i, integrationTestTimeout)
	defer done()
	for {
		err := assertion(t, i)
		if err == nil {
			return
		}
		select {
		case <-waitCtx.Done():
			t.Fatalf("Timed out waiting for assertion to pass: %s", err)

		case <-time.After(time.Millisecond * 200):
		}
	}
}

// startProcess runs a binary in the background.
func startProcess(t testing.TB, ctx context.Context, args ...string) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.Command(ctx, log.Info, "..", args[0], args[1:]...)
	err := cmd.Start()
	assert.NoError(t, err)
	terminated := make(chan bool)
	go func() {
		err := cmd.Wait()
		select {
		case <-terminated:
		default:
			cancel()
			assert.NoError(t, err)
		}
	}()
	t.Cleanup(func() {
		close(terminated)
		err := cmd.Kill(syscall.SIGTERM)
		assert.NoError(t, err)
		cancel()
	})
	return ctx
}

type logWriter struct {
	logger interface{ Log(...any) }
	buffer []byte
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	for {
		index := bytes.IndexByte(p, '\n')
		if index == -1 {
			l.buffer = append(l.buffer, p...)
			return n, nil
		} else {
			l.buffer = append(l.buffer, p[:index]...)
			l.logger.Log(string(l.buffer))
			l.buffer = l.buffer[:0]
			p = p[index+1:]
		}
	}
}
