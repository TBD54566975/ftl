//go:build integration

package simple_test

import (
	"bytes"
	"context"
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
	"github.com/otiai10/copy"

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

func infof(format string, args ...any) {
	fmt.Printf("\033[32m\033[1mINFO: "+format+"\033[0m\n", args...)
}

var buildOnce sync.Once

// run an integration test.
func run(t *testing.T, ftlConfigPath string, actions ...action) {
	tmpDir := t.TempDir()

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	rootDir := internal.GitRoot("")
	testData := filepath.Join(cwd, "testdata", "go")

	if ftlConfigPath != "" {
		tmpConfigPath := filepath.Join(tmpDir, ftlConfigPath)
		infof("Copying %s to %s", filepath.Join(testData, ftlConfigPath), tmpConfigPath)
		copy.Copy(filepath.Join(testData, ftlConfigPath), tmpConfigPath)
		t.Setenv("FTL_CONFIG", tmpConfigPath)
	}

	// Build FTL binary
	logger := log.Configure(&logWriter{logger: t}, log.Config{Level: log.Debug})
	ctx := log.ContextWithLogger(context.Background(), logger)
	binDir := filepath.Join(rootDir, "build", "release")

	buildOnce.Do(func() {
		infof("Building ftl")
		err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build", "ftl").RunBuffered(ctx)
		assert.NoError(t, err)
	})

	controller := rpc.Dial(ftlv1connect.NewControllerServiceClient, "http://localhost:8892", log.Debug)
	verbs := rpc.Dial(ftlv1connect.NewVerbServiceClient, "http://localhost:8892", log.Debug)

	infof("Starting ftl cluster")
	ctx = startProcess(ctx, t, filepath.Join(binDir, "ftl"), "serve", "--recreate")

	ic := testContext{
		Context:    ctx,
		rootDir:    rootDir,
		testData:   testData,
		workDir:    tmpDir,
		binDir:     binDir,
		controller: controller,
		verbs:      verbs,
	}

	infof("Waiting for controller to be ready")
	ic.AssertWithRetry(t, func(t testing.TB, ic testContext) error {
		_, err := ic.controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
		return err
	})

	infof("Starting test")

	for _, action := range actions {
		ic.AssertWithRetry(t, action)
	}
}

type testContext struct {
	context.Context
	// Temporary directory the test is executing in.
	workDir string
	// Root of FTL repo.
	rootDir string
	// Path to testdata directory for the current language.
	testData string
	// Path to the "bin" directory.
	binDir string

	controller ftlv1connect.ControllerServiceClient
	verbs      ftlv1connect.VerbServiceClient
}

// AssertWithRetry asserts that the given action passes within the timeout.
func (i testContext) AssertWithRetry(t testing.TB, assertion action) {
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

type action func(t testing.TB, ic testContext) error

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
