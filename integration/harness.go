//go:build integration

package integration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal"
	ftlexec "github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

var integrationTestTimeout = func() time.Duration {
	timeout := optional.Zero(os.Getenv("FTL_INTEGRATION_TEST_TIMEOUT")).Default("5s")
	d, err := time.ParseDuration(timeout)
	if err != nil {
		panic(err)
	}
	return d
}()

func Infof(format string, args ...any) {
	fmt.Printf("\033[32m\033[1mINFO: "+format+"\033[0m\n", args...)
}

var buildOnce sync.Once

// Run an integration test.
// ftlConfigPath: if FTL_CONFIG should be set for this test, then pass in the relative
//
//	path based on ./testdata/go/ where "." denotes the directory containing the
//	integration test (e.g. for "integration/harness_test.go" supplying
//	"database/ftl-project.toml" would set FTL_CONFIG to
//	"integration/testdata/go/database/ftl-project.toml").
func Run(t *testing.T, ftlConfigPath string, actions ...Action) {
	tmpDir := t.TempDir()

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	rootDir, ok := internal.GitRoot("").Get()
	assert.True(t, ok)

	if ftlConfigPath != "" {
		// Use a path into the testdata directory instead of one relative to
		// tmpDir. Otherwise we have a chicken and egg situation where the config
		// can't be loaded until the module is copied over, and the config itself
		// is used by FTL during startup.
		t.Setenv("FTL_CONFIG", filepath.Join(cwd, "testdata", "go", ftlConfigPath))
	}

	// Build FTL binary
	logger := log.Configure(&logWriter{logger: t}, log.Config{Level: log.Debug})
	ctx := log.ContextWithLogger(context.Background(), logger)
	binDir := filepath.Join(rootDir, "build", "release")

	buildOnce.Do(func() {
		Infof("Building ftl")
		err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build", "ftl").RunBuffered(ctx)
		assert.NoError(t, err)
	})

	controller := rpc.Dial(ftlv1connect.NewControllerServiceClient, "http://localhost:8892", log.Debug)
	verbs := rpc.Dial(ftlv1connect.NewVerbServiceClient, "http://localhost:8892", log.Debug)

	Infof("Starting ftl cluster")
	ctx = startProcess(ctx, t, filepath.Join(binDir, "ftl"), "serve", "--recreate")

	ic := TestContext{
		Context:    ctx,
		rootDir:    rootDir,
		testData:   filepath.Join(cwd, "testdata", "go"),
		workDir:    tmpDir,
		binDir:     binDir,
		Controller: controller,
		Verbs:      verbs,
	}

	Infof("Waiting for controller to be ready")
	ic.AssertWithRetry(t, func(t testing.TB, ic TestContext) {
		_, err := ic.Controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
		assert.NoError(t, err)
	})

	Infof("Starting test")

	for _, action := range actions {
		ic.AssertWithRetry(t, action)
	}
}

type TestContext struct {
	context.Context
	// Temporary directory the test is executing in.
	workDir string
	// Root of FTL repo.
	rootDir string
	// Path to testdata directory for the current language.
	testData string
	// Path to the "bin" directory.
	binDir string

	Controller ftlv1connect.ControllerServiceClient
	Verbs      ftlv1connect.VerbServiceClient
}

// AssertWithRetry asserts that the given action passes within the timeout.
func (i TestContext) AssertWithRetry(t testing.TB, assertion Action) {
	waitCtx, done := context.WithTimeout(i, integrationTestTimeout)
	defer done()
	for {
		err := i.runAssertionOnce(t, assertion)
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

// Run an assertion, wrapping testing.TB in an implementation that panics on failure, propagating the error.
func (i TestContext) runAssertionOnce(t testing.TB, assertion Action) (err error) {
	defer func() {
		switch r := recover().(type) {
		case TestingError:
			err = errors.New(string(r))

		case nil:
			return

		default:
			panic(r)
		}
	}()
	assertion(T{t}, i)
	return nil
}

type Action func(t testing.TB, ic TestContext)

type logWriter struct {
	mu     sync.Mutex
	logger interface{ Log(...any) }
	buffer []byte
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for {
		index := bytes.IndexByte(p, '\n')
		if index == -1 {
			l.buffer = append(l.buffer, p...)
			return n, nil
		}
		l.buffer = append(l.buffer, p[:index]...)
		l.logger.Log(string(l.buffer))
		l.buffer = l.buffer[:0]
		p = p[index+1:]
	}
}

// startProcess runs a binary in the background.
func startProcess(ctx context.Context, t testing.TB, args ...string) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(ctx)
	cmd := ftlexec.Command(ctx, log.Debug, "..", args[0], args[1:]...)
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
