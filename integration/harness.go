//go:build integration

package integration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"syscall"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/otiai10/copy"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal"
	ftlexec "github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

func integrationTestTimeout() time.Duration {
	timeout := optional.Zero(os.Getenv("FTL_INTEGRATION_TEST_TIMEOUT")).Default("5s")
	d, err := time.ParseDuration(timeout)
	if err != nil {
		panic(err)
	}
	return d
}

func Infof(format string, args ...any) {
	fmt.Printf("\033[32m\033[1mINFO: "+format+"\033[0m\n", args...)
}

var buildOnce sync.Once

// An Option for configuring the integration test harness.
type Option func(*options)

// ActionOrOption is a type that can be either an Action or an Option.
type ActionOrOption any

// WithLanguages is a Run* option that specifies the languages to test.
//
// Defaults to "go" if not provided.
func WithLanguages(languages ...string) Option {
	return func(o *options) {
		o.languages = languages
	}
}

// WithFTLConfig is a Run* option that specifies the FTL config to use.
//
// This will set FTL_CONFIG for this test, then pass in the relative
// path based on ./testdata/go/ where "." denotes the directory containing the
// integration test (e.g. for "integration/harness_test.go" supplying
// "database/ftl-project.toml" would set FTL_CONFIG to
// "integration/testdata/go/database/ftl-project.toml").
func WithFTLConfig(path string) Option {
	return func(o *options) {
		o.ftlConfigPath = path
	}
}

// WithEnvar is a Run* option that specifies an environment variable to set.
func WithEnvar(key, value string) Option {
	return func(o *options) {
		o.envars[key] = value
	}
}

// BuildJava is a Run* option that ensures the Java runtime is built.
// If the test languages contain java this is not necessary, as it is implied
func BuildJava() Option {
	return func(o *options) {
		o.requireJava = true
	}
}

// WithoutController is a Run* option that disables starting the controller.
func WithoutController() Option {
	return func(o *options) {
		o.startController = false
	}
}

type options struct {
	languages       []string
	ftlConfigPath   string
	startController bool
	requireJava     bool
	envars          map[string]string
}

// Run an integration test.
func Run(t *testing.T, actionsOrOptions ...ActionOrOption) {
	run(t, actionsOrOptions...)
}

func run(t *testing.T, actionsOrOptions ...ActionOrOption) {
	opts := options{
		startController: true,
		languages:       []string{"go"},
		envars:          map[string]string{},
	}
	actions := []Action{}
	for _, opt := range actionsOrOptions {
		switch o := opt.(type) {
		case Action:
			actions = append(actions, o)

		case func(t testing.TB, ic TestContext):
			actions = append(actions, Action(o))

		case Option:
			o(&opts)

		case func(*options):
			o(&opts)

		default:
			panic(fmt.Sprintf("expected Option or Action, not %T", opt))
		}
	}

	for key, value := range opts.envars {
		t.Setenv(key, value)
	}

	tmpDir := t.TempDir()

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	rootDir, ok := internal.GitRoot("").Get()
	assert.True(t, ok)

	if opts.ftlConfigPath != "" {
		// TODO: We shouldn't be copying the shared config from the "go" testdata...
		opts.ftlConfigPath = filepath.Join(cwd, "testdata", "go", opts.ftlConfigPath)
		projectPath := filepath.Join(tmpDir, "ftl-project.toml")

		// Copy the specified FTL config to the temporary directory.
		err = copy.Copy(opts.ftlConfigPath, projectPath)
		if err == nil {
			t.Setenv("FTL_CONFIG", projectPath)
		} else {
			// Use a path into the testdata directory instead of one relative to
			// tmpDir. Otherwise we have a chicken and egg situation where the config
			// can't be loaded until the module is copied over, and the config itself
			// is used by FTL during startup.
			// Some tests still rely on this behavior, so we can't remove it entirely.
			t.Logf("Failed to copy %s to %s: %s", opts.ftlConfigPath, projectPath, err)
			t.Setenv("FTL_CONFIG", opts.ftlConfigPath)
		}

	} else {
		err = os.WriteFile(filepath.Join(tmpDir, "ftl-project.toml"), []byte(`name = "integration"`), 0644)
		assert.NoError(t, err)
	}

	// Build FTL binary
	logger := log.Configure(&logWriter{logger: t}, log.Config{Level: log.Debug})
	ctx := log.ContextWithLogger(context.Background(), logger)
	binDir := filepath.Join(rootDir, "build", "release")

	buildOnce.Do(func() {
		Infof("Building ftl")
		err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build", "ftl").RunBuffered(ctx)
		assert.NoError(t, err)
		if opts.requireJava || slices.Contains(opts.languages, "java") {
			err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build-java", "-DskipTests").RunBuffered(ctx)
			assert.NoError(t, err)
		}
	})

	for _, language := range opts.languages {
		ctx, done := context.WithCancel(ctx)
		t.Run(language, func(t *testing.T) {
			verbs := rpc.Dial(ftlv1connect.NewVerbServiceClient, "http://localhost:8892", log.Debug)

			var controller ftlv1connect.ControllerServiceClient
			var console pbconsoleconnect.ConsoleServiceClient
			if opts.startController {
				controller = rpc.Dial(ftlv1connect.NewControllerServiceClient, "http://localhost:8892", log.Debug)
				console = rpc.Dial(pbconsoleconnect.NewConsoleServiceClient, "http://localhost:8892", log.Debug)

				Infof("Starting ftl cluster")
				ctx = startProcess(ctx, t, filepath.Join(binDir, "ftl"), "serve", "--recreate")
			}

			ic := TestContext{
				Context:  ctx,
				RootDir:  rootDir,
				testData: filepath.Join(cwd, "testdata", language),
				workDir:  tmpDir,
				binDir:   binDir,
				Verbs:    verbs,
				realT:    t,
				language: language,
			}

			if opts.startController {
				ic.Controller = controller
				ic.Console = console

				Infof("Waiting for controller to be ready")
				ic.AssertWithRetry(t, func(t testing.TB, ic TestContext) {
					_, err := ic.Controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
					assert.NoError(t, err)
				})
			}

			Infof("Starting test")

			for _, action := range actions {
				ic.AssertWithRetry(t, action)
			}
		})
		done()
	}
}

type TestContext struct {
	context.Context
	// Temporary directory the test is executing in.
	workDir string
	// Root of FTL repo.
	RootDir string
	// Path to testdata directory for the current language.
	testData string
	// Path to the "bin" directory.
	binDir string
	// The Language under test
	language string

	Controller ftlv1connect.ControllerServiceClient
	Console    pbconsoleconnect.ConsoleServiceClient
	Verbs      ftlv1connect.VerbServiceClient

	realT *testing.T
}

func (i TestContext) Run(name string, f func(t *testing.T)) bool {
	return i.realT.Run(name, f)
}

// WorkingDir returns the temporary directory the test is executing in.
func (i TestContext) WorkingDir() string { return i.workDir }

// AssertWithRetry asserts that the given action passes within the timeout.
func (i TestContext) AssertWithRetry(t testing.TB, assertion Action) {
	waitCtx, done := context.WithTimeout(i, integrationTestTimeout())
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
			fmt.Println(string(r))

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

type SubTest struct {
	Name   string
	Action Action
}

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

// startProcess runs a binary in the background and terminates it when the test completes.
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
