//go:build integration || infrastructure || smoketest

package integration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"syscall"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/otiai10/copy"
	"k8s.io/client-go/kubernetes"

	kubecore "k8s.io/api/core/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/yaml"

	"github.com/TBD54566975/ftl/backend/controller/scaling/k8sscaling"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/internal"
	ftlexec "github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

const dumpPath = "/tmp/ftl-kube-report"

func (i TestContext) integrationTestTimeout() time.Duration {
	timeout := optional.Zero(os.Getenv("FTL_INTEGRATION_TEST_TIMEOUT")).Default("5s")
	d, err := time.ParseDuration(timeout)
	if err != nil {
		panic(err)
	}
	if i.kubeClient.Ok() {
		// kube can be slow, give it some time
		return d * 5
	}
	return d
}

func Infof(format string, args ...any) {
	fmt.Printf("\033[32m\033[1mINFO: "+format+"\033[0m\n", args...)
}

var buildOnce sync.Once

var buildOnceOptions *options

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

// WithKubernetes is a Run* option that specifies tests should be run on a kube cluster
func WithKubernetes() Option {
	return func(o *options) {
		o.kube = true
		o.startController = false
	}
}

// WithLocalstack is a Run* option that specifies tests should be run on a localstack container
func WithLocalstack() Option {
	return func(o *options) {
		o.localstack = true
	}
}

// WithConsole is a Run* option that specifies tests should build and start the console
func WithConsole() Option {
	return func(o *options) {
		o.console = true
	}
}

// WithTestDataDir sets the directory from which to look for test data.
//
// Defaults to "testdata/<language>" if not provided.
func WithTestDataDir(dir string) Option {
	return func(o *options) {
		o.testDataDir = dir
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

// WithJavaBuild is a Run* option that ensures the Java runtime is built.
// If the test languages contain java this is not necessary, as it is implied
// Note that this will not actually add Java as a language under test
func WithJavaBuild() Option {
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

// WithProvisioner is a Run* option that starts the provisioner service.
// if set, all deployments are done through the provisioner
func WithProvisioner() Option {
	return func(o *options) {
		o.startProvisioner = true
		// provisioner always needs a controller to talk to
		o.startController = true
	}
}

// WithProvisionerConfig is a Run* option that specifies the provisioner config to use.
func WithProvisionerConfig(config string) Option {
	return func(o *options) {
		o.provisionerConfig = config
	}
}

type options struct {
	languages         []string
	testDataDir       string
	ftlConfigPath     string
	startController   bool
	startProvisioner  bool
	provisionerConfig string
	requireJava       bool
	envars            map[string]string
	kube              bool
	localstack        bool
	console           bool
}

// Run an integration test.
func Run(t *testing.T, actionsOrOptions ...ActionOrOption) {
	t.Helper()
	run(t, actionsOrOptions...)
}

func run(t *testing.T, actionsOrOptions ...ActionOrOption) {
	t.Helper()
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

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	rootDir, ok := internal.GitRoot("").Get()
	assert.True(t, ok)

	// Build FTL binary
	logger := log.Configure(&logWriter{logger: t}, log.Config{Level: log.Debug})
	ctx := log.ContextWithLogger(context.Background(), logger)
	binDir := filepath.Join(rootDir, "build", "release")

	var kubeClient *kubernetes.Clientset
	var kubeNamespace string
	buildOnce.Do(func() {
		buildOnceOptions = &opts
		if opts.kube {
			// This command will build a linux/amd64 version of FTL and deploy it to the kube cluster
			Infof("Building FTL and deploying to kube")
			err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "chart", "dep-update").RunBuffered(ctx)
			assert.NoError(t, err)
			err = ftlexec.Command(ctx, log.Debug, filepath.Join(rootDir, "deployment"), "just", "setup-cluster").RunBuffered(ctx)
			assert.NoError(t, err)
			err = ftlexec.Command(ctx, log.Debug, filepath.Join(rootDir, "deployment"), "just", "install-istio").RunBuffered(ctx)
			assert.NoError(t, err)

			skipKubeFullDeploy := os.Getenv("SKIP_KUBE_FULL_DEPLOY") == "true"
			if skipKubeFullDeploy {
				Infof("Skipping full deploy since SKIP_KUBE_FULL_DEPLOY is set")
			} else {
				err = ftlexec.Command(ctx, log.Debug, filepath.Join(rootDir, "deployment"), "just", "full-deploy").RunBuffered(ctx)
				assert.NoError(t, err)
			}
			if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
				// If we are already on linux/amd64 we don't need to rebuild, otherwise we now need a native one to interact with the kube cluster
				Infof("Building FTL for native OS")
				err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build", "ftl").RunBuffered(ctx)
				assert.NoError(t, err)
			}
			kubeClient, err = k8sscaling.CreateClientSet()
			assert.NoError(t, err)
			kubeNamespace, err = k8sscaling.GetCurrentNamespace()
			assert.NoError(t, err)
			// We create the client before, as kube itself is up, we are just waiting for the deployments
			err = ftlexec.Command(ctx, log.Debug, filepath.Join(rootDir, "deployment"), "just", "wait-for-kube").RunBuffered(ctx)
			if err != nil {
				dumpKubePods(ctx, optional.Ptr(kubeClient), kubeNamespace)
			}
			assert.NoError(t, err)
		} else if opts.console {
			Infof("Building ftl with console")
			err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build", "ftl").RunBuffered(ctx)
			assert.NoError(t, err)
		} else {
			Infof("Building ftl without console")
			err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build-without-frontend", "ftl").RunBuffered(ctx)
			assert.NoError(t, err)
		}
		if opts.requireJava || slices.Contains(opts.languages, "java") || slices.Contains(opts.languages, "kotlin") {
			err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "build-jvm", "-DskipTests", "-B").Run()
			assert.NoError(t, err)
		}
		if opts.localstack {
			err = ftlexec.Command(ctx, log.Debug, rootDir, "just", "localstack").RunBuffered(ctx)
			assert.NoError(t, err)
		}
	})

	assert.Equal(t, *buildOnceOptions, opts, "Options changed between test runs")

	for _, language := range opts.languages {
		ctx, done := context.WithCancel(ctx)
		t.Run(language, func(t *testing.T) {
			t.Helper()
			tmpDir := initWorkDir(t, cwd, opts)

			verbs := rpc.Dial(ftlv1connect.NewVerbServiceClient, "http://localhost:8892", log.Debug)

			var controller ftlv1connect.ControllerServiceClient
			var console pbconsoleconnect.ConsoleServiceClient
			var provisioner provisionerconnect.ProvisionerServiceClient
			if opts.startController {
				Infof("Starting ftl cluster")

				args := []string{filepath.Join(binDir, "ftl"), "serve", "--recreate"}
				if !opts.console {
					args = append(args, "--no-console")
				}
				if opts.startProvisioner {
					args = append(args, "--provisioners=1")

					if opts.provisionerConfig != "" {
						configFile := filepath.Join(tmpDir, "provisioner-plugin-config.toml")
						os.WriteFile(configFile, []byte(opts.provisionerConfig), 0644)
						args = append(args, "--provisioner-plugin-config="+configFile)
					}
				}
				ctx = startProcess(ctx, t, args...)
			}
			if opts.startController || opts.kube {
				controller = rpc.Dial(ftlv1connect.NewControllerServiceClient, "http://localhost:8892", log.Debug)
				console = rpc.Dial(pbconsoleconnect.NewConsoleServiceClient, "http://localhost:8892", log.Debug)
			}
			if opts.startProvisioner {
				provisioner = rpc.Dial(provisionerconnect.NewProvisionerServiceClient, "http://localhost:8893", log.Debug)
			}

			testData := filepath.Join(cwd, "testdata", language)
			if opts.testDataDir != "" {
				testData = opts.testDataDir
			}

			ic := TestContext{
				Context:       ctx,
				RootDir:       rootDir,
				testData:      testData,
				workDir:       tmpDir,
				binDir:        binDir,
				Verbs:         verbs,
				realT:         t,
				Language:      language,
				kubeNamespace: kubeNamespace,
				kubeClient:    optional.Ptr(kubeClient),
			}
			defer dumpKubePods(ctx, ic.kubeClient, ic.kubeNamespace)

			if opts.startController || opts.kube {
				ic.Controller = controller
				ic.Console = console

				Infof("Waiting for controller to be ready")
				ic.AssertWithRetry(t, func(t testing.TB, ic TestContext) {
					_, err := ic.Controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
					assert.NoError(t, err)
				})
			}

			if opts.startProvisioner {
				ic.Provisioner = provisioner

				Infof("Waiting for provisioner to be ready")
				ic.AssertWithRetry(t, func(t testing.TB, ic TestContext) {
					_, err := ic.Provisioner.Ping(ic, connect.NewRequest(&ftlv1.PingRequest{}))
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

func initWorkDir(t testing.TB, cwd string, opts options) string {
	tmpDir := t.TempDir()

	if opts.ftlConfigPath != "" {
		// TODO: We shouldn't be copying the shared config from the "go" testdata...
		opts.ftlConfigPath = filepath.Join(cwd, "testdata", "go", opts.ftlConfigPath)
		projectPath := filepath.Join(tmpDir, "ftl-project.toml")

		// Copy the specified FTL config to the temporary directory.
		err := copy.Copy(opts.ftlConfigPath, projectPath)
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
		err := os.WriteFile(filepath.Join(tmpDir, "ftl-project.toml"), []byte(`name = "integration"`), 0644)
		assert.NoError(t, err)
	}
	return tmpDir
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
	Language string
	// Set if the test is running on kubernetes
	kubeClient    optional.Option[kubernetes.Clientset]
	kubeNamespace string

	Controller  ftlv1connect.ControllerServiceClient
	Provisioner provisionerconnect.ProvisionerServiceClient
	Console     pbconsoleconnect.ConsoleServiceClient
	Verbs       ftlv1connect.VerbServiceClient

	realT *testing.T
}

func (i TestContext) Run(name string, f func(t *testing.T)) bool {
	return i.realT.Run(name, f)
}

// WorkingDir returns the temporary directory the test is executing in.
func (i TestContext) WorkingDir() string { return i.workDir }

// AssertWithRetry asserts that the given action passes within the timeout.
func (i TestContext) AssertWithRetry(t testing.TB, assertion Action) {
	t.Helper()
	waitCtx, done := context.WithTimeout(i, i.integrationTestTimeout())
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
	t.Helper()
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

func dumpKubePods(ctx context.Context, kubeClient optional.Option[kubernetes.Clientset], kubeNamespace string) {
	if client, ok := kubeClient.Get(); ok {
		_ = os.RemoveAll(dumpPath) // #nosec
		list, err := client.CoreV1().Pods(kubeNamespace).List(ctx, kubemeta.ListOptions{})
		if err == nil {
			for _, pod := range list.Items {
				Infof("Dumping logs for pod %s", pod.Name)
				podPath := filepath.Join(dumpPath, pod.Name)
				err := os.MkdirAll(podPath, 0755) // #nosec
				if err != nil {
					Infof("Error creating directory %s: %v", podPath, err)
					continue
				}
				podYaml, err := yaml.Marshal(pod)
				if err != nil {
					Infof("Error marshalling pod %s: %v", pod.Name, err)
					continue
				}
				err = os.WriteFile(filepath.Join(podPath, "pod.yaml"), podYaml, 0644) // #nosec
				if err != nil {
					Infof("Error writing pod %s: %v", pod.Name, err)
					continue
				}
				for _, container := range pod.Spec.Containers {
					path := filepath.Join(dumpPath, pod.Name, container.Name+".log")
					req := client.CoreV1().Pods(kubeNamespace).GetLogs(pod.Name, &kubecore.PodLogOptions{Container: container.Name})
					podLogs, err := req.Stream(context.Background())
					if err != nil {
						Infof("Error getting logs for pod %s: %v", pod.Name, err)
						continue
					}
					defer func() {
						_ = podLogs.Close()
					}()
					buf := new(bytes.Buffer)
					_, err = io.Copy(buf, podLogs)
					if err != nil {
						Infof("Error copying logs for pod %s: %v", pod.Name, err)
						continue
					}
					str := buf.String()
					err = os.WriteFile(path, []byte(str), 0644) // #nosec
					if err != nil {
						Infof("Error writing logs for pod %s: %v", pod.Name, err)
					}
				}
			}
		}
	}
}
