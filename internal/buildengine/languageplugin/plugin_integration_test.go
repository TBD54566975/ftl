//go:build integration

package languageplugin

import (
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/result"
	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/sync/errgroup"

	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/flock"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/internal/watch"
)

// These integration tests are meant as a test suite for external plugins.
//
// It is not meant to mirror exactly how FTL calls external plugins, but rather
// to test plugins against the langauge plugin protocol.
//
// It would be great to allow externally built plugins to be tested against this
// test suite.
//
// Each language must provide a module with the name plugintest, which has:
// - A verb which includes in its name the string VERB_NAME_SNIPPET (case sensitive), eg: "func Verbaabbcc(...)"
// - No dependencies, but has lines that include the string "uncommentForDependency:", which enable a
//     on dependency on the module "dependable" when everything before the colon is removed.
//     - dependable.Data is a data type which can be used to avoid unused dependency warnings.

var client *pluginClientImpl
var bindURL *url.URL
var config moduleconfig.ModuleConfig
var buildChan chan result.Result[*langpb.BuildEvent]
var buildChanCancel streamCancelFunc

const MODULE_NAME = "plugintest"
const VERB_NAME_SNIPPET = "aabbcc"

type BuildResultType int

const (
	SUCCESS BuildResultType = iota
	FAILURE
	SUCCESSORFAILURE
)

func TestBuilds(t *testing.T) {
	sch := generateInitialSchema(t)
	in.Run(t,
		in.WithLanguages("go", "java"),
		in.WithoutController(),
		in.CopyModule(MODULE_NAME),
		startPlugin(),
		setUpModuleConfig(MODULE_NAME),
		generateStubs(sch.Modules...),
		syncStubReferences("builtin", "dependable"),

		// Build once
		build(false, []string{}, sch, "build-once"),
		waitForBuildToEnd(SUCCESS, "build-once", false, nil),

		// Sending build context updates should fail if the plugin has no way to send back a build result
		in.Fail(
			sendUpdatedBuildContext("no-build-stream", []string{}, sch),
			"expected error when sending build context update without a build stream",
		),

		// Build and enable rebuilding automatically
		build(true, []string{}, sch, "build-and-watch"),
		waitForBuildToEnd(SUCCESS, "build-and-watch", false, nil),

		// Update verb name and expect auto rebuild started and ended
		modifyVerbName(MODULE_NAME, VERB_NAME_SNIPPET, "aaabbbccc"),
		waitForAutoRebuildToStart("build-and-watch"),
		waitForBuildToEnd(SUCCESS, "build-and-watch", true, func(t testing.TB, ic in.TestContext, event *langpb.BuildEvent) {
			successEvent, ok := event.Event.(*langpb.BuildEvent_BuildSuccess)
			assert.True(t, ok)
			_, found := slices.Find(successEvent.BuildSuccess.Module.Decls, func(decl *schemapb.Decl) bool {
				verb, ok := decl.Value.(*schemapb.Decl_Verb)
				if !ok {
					return false
				}
				return strings.Contains(verb.Verb.Name, "aaabbbccc")
			})
			assert.True(t, found, "expected verb name to be updated to include %q", "aaabbbccc")
		}),

		// Trigger an auto rebuild, but when we are told of the build being started, send a build context update
		// to force a new build
		modifyVerbName(MODULE_NAME, "aaabbbccc", "aaaabbbbcccc"),
		waitForAutoRebuildToStart("build-and-watch"),
		sendUpdatedBuildContext("explicit-build", []string{}, sch),
		waitForBuildToEnd(SUCCESSORFAILURE, "build-and-watch", true, nil),
		waitForBuildToEnd(SUCCESS, "explicit-build", false, nil),

		// Trigger 2 explicit builds, make sure we get a response for both of them (first one can fail)
		sendUpdatedBuildContext("double-build-1", []string{}, sch),
		sendUpdatedBuildContext("double-build-2", []string{}, sch),
		waitForBuildToEnd(SUCCESSORFAILURE, "double-build-1", false, nil),
		waitForBuildToEnd(SUCCESS, "double-build-2", false, nil),

		killPlugin(),
	)
}

func TestDependenciesUpdate(t *testing.T) {
	sch := generateInitialSchema(t)

	in.Run(t,
		in.WithLanguages("go"), //no java support yet, as it relies on writeGenericSchemaFiles
		in.WithoutController(),
		in.CopyModule(MODULE_NAME),
		startPlugin(),
		setUpModuleConfig(MODULE_NAME),
		generateStubs(sch.Modules...),
		syncStubReferences("builtin", "dependable"),

		// Build
		build(false, []string{}, sch, "initial-ctx"),
		waitForBuildToEnd(SUCCESS, "initial-ctx", false, nil),

		// Add dependency, build, and expect a failure due to invalidated dependencies
		addDependency(MODULE_NAME, "dependable"),
		build(false, []string{}, sch, "detect-dep"),
		waitForBuildToEnd(FAILURE, "detect-dep", false, func(t testing.TB, ic in.TestContext, event *langpb.BuildEvent) {
			failureEvent, ok := event.Event.(*langpb.BuildEvent_BuildFailure)
			assert.True(t, ok)
			assert.True(t, failureEvent.BuildFailure.InvalidateDependencies, "expected dependencies to be invalidated")
		}),

		// Build with new dependency
		build(false, []string{"dependable"}, sch, "dep-added"),
		waitForBuildToEnd(SUCCESS, "dep-added", false, nil),

		killPlugin(),
	)
}

// TestBuildLock tests that the build lock file is created and removed as expected for each build.
func TestBuildLock(t *testing.T) {
	sch := generateInitialSchema(t)

	in.Run(t,
		in.WithLanguages("go", "java"),
		in.WithoutController(),
		in.CopyModule(MODULE_NAME),
		startPlugin(),
		setUpModuleConfig(MODULE_NAME),
		generateStubs(sch.Modules...),
		syncStubReferences("builtin", "dependable"),

		// Build and enable rebuilding automatically
		checkBuildLockLifecycle(
			build(true, []string{}, sch, "build-and-watch"),
			waitForBuildToEnd(SUCCESS, "build-and-watch", false, nil),
		),

		// Update verb name and expect auto rebuild started and ended
		modifyVerbName(MODULE_NAME, VERB_NAME_SNIPPET, "aaabbbccc"),
		checkBuildLockLifecycle(
			waitForAutoRebuildToStart("build-and-watch"),
			waitForBuildToEnd(SUCCESS, "build-and-watch", true, nil),
		),
	)
}

// TestBuildsWhenAlreadyLocked tests how builds work if there are locks already present.
func TestBuildsWhenAlreadyLocked(t *testing.T) {
	sch := generateInitialSchema(t)

	in.Run(t,
		in.WithLanguages("go", "java"),
		in.WithoutController(),
		in.CopyModule(MODULE_NAME),
		startPlugin(),
		setUpModuleConfig(MODULE_NAME),
		generateStubs(sch.Modules...),
		syncStubReferences("builtin", "dependable"),

		// Build and enable rebuilding automatically
		checkBuildLockLifecycle(
			build(true, []string{}, sch, "build-and-watch"),
			waitForBuildToEnd(SUCCESS, "build-and-watch", false, nil),
		),

		// Confirm that build lock changes do not trigger a rebuild triggered by file changes
		obtainAndReleaseBuildLock(3*time.Second),
		checkForNoEvents(3*time.Second),

		// Confirm that builds fail or stall when a lock file is already present
		checkLockedBehavior(
			sendUpdatedBuildContext("updated-ctx", []string{}, sch),
			waitForBuildToEnd(FAILURE, "updated-ctx", false, nil),
		),
	)
}

func generateInitialSchema(t *testing.T) *schema.Schema {
	t.Helper()

	sch, err := schema.ValidateSchema(&schema.Schema{
		Modules: []*schema.Module{
			{
				Name: "dependable",
				Decls: []schema.Decl{
					&schema.Data{
						Name:   "Data",
						Export: true,
					},
				},
			},
		},
	})
	assert.NoError(t, err)
	return sch
}

func startPlugin() in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Starting plugin")
		baseBind, err := url.Parse("http://127.0.0.1:8893")
		assert.NoError(t, err)
		bindAllocator, err := bind.NewBindAllocator(baseBind, 0)
		assert.NoError(t, err)

		bindURL, err = bindAllocator.Next()
		assert.NoError(t, err)
		client, err = newClientImpl(ic.Context, ic.WorkingDir(), ic.Language, "test")
		assert.NoError(t, err)
	}
}

func killPlugin() in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Killing plugin")
		err := client.kill()
		assert.NoError(t, err, "could not kill plugin")

		time.Sleep(1 * time.Second)

		// check that the bind port is freed (ie: the plugin has exited)
		var l *net.TCPListener
		_, portStr, err := net.SplitHostPort(bindURL.Host)
		assert.NoError(t, err)
		port, err := strconv.Atoi(portStr)
		assert.NoError(t, err)
		l, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(bindURL.Hostname()), Port: port})
		if err != nil {
			// panic so that we don't retry, which can hide the real error
			panic("plugin's port is still in use")
		}
		_ = l.Close()
	}
}

func setUpModuleConfig(moduleName string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Setting up config for %s", moduleName)
		path := filepath.Join(ic.WorkingDir(), moduleName)
		resp, err := client.moduleConfigDefaults(ic.Context, connect.NewRequest(&langpb.ModuleConfigDefaultsRequest{
			Dir: path,
		}))
		assert.NoError(t, err)

		unvalidatedConfig, err := moduleconfig.LoadConfig(path)
		assert.NoError(t, err)

		config, err = unvalidatedConfig.FillDefaultsAndValidate(customDefaultsFromProto(resp.Msg))
		assert.NoError(t, err)
	}
}

func generateStubs(moduleSchs ...*schema.Module) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Generating stubs for %v", slices.Map(moduleSchs, func(m *schema.Module) string { return m.Name }))
		wg, wgctx := errgroup.WithContext(ic.Context)
		for _, moduleSch := range moduleSchs {
			wg.Go(func() error {
				var configForStub moduleconfig.ModuleConfig
				if moduleSch.Name == config.Module {
					configForStub = config
				} else if moduleSch.Name == "builtin" {
					configForStub = moduleconfig.ModuleConfig{
						Module:   "builtin",
						Language: "go",
					}
				} else {
					configForStub = moduleconfig.ModuleConfig{
						Module:   moduleSch.Name,
						Language: "fake",
					}
				}

				configForStubProto, err := langpb.ModuleConfigToProto(configForStub.Abs())
				assert.NoError(t, err)

				var nativeConfigProto *langpb.ModuleConfig
				if moduleSch.Name != config.Module {
					nativeConfigProto, err = langpb.ModuleConfigToProto(config.Abs())
					assert.NoError(t, err)
				}

				path := filepath.Join(ic.WorkingDir(), ".ftl", config.Language, "modules", configForStub.Module)
				err = os.MkdirAll(path, 0750)
				assert.NoError(t, err)

				moduleProto, ok := moduleSch.ToProto().(*schemapb.Module) //nolint:forcetypeassert
				assert.True(t, ok)
				_, err = client.generateStubs(wgctx, connect.NewRequest(&langpb.GenerateStubsRequest{
					Dir:                path,
					Module:             moduleProto,
					ModuleConfig:       configForStubProto,
					NativeModuleConfig: nativeConfigProto,
				}))
				assert.NoError(t, err)
				return nil
			})
		}
		err := wg.Wait()
		assert.NoError(t, err)
	}
}

func syncStubReferences(moduleNames ...string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Syncing stub references for %v", moduleNames)

		configProto, err := langpb.ModuleConfigToProto(config.Abs())
		assert.NoError(t, err)

		_, err = client.syncStubReferences(ic.Context, connect.NewRequest(&langpb.SyncStubReferencesRequest{
			ModuleConfig: configProto,
			StubsRoot:    filepath.Join(ic.WorkingDir(), ".ftl", config.Language, "modules"),
			Modules:      moduleNames,
		}))
		assert.NoError(t, err)
	}
}

func build(rebuildAutomatically bool, dependencies []string, sch *schema.Schema, contextId string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Plugin building: %s", contextId)
		configProto, err := langpb.ModuleConfigToProto(config.Abs())
		assert.NoError(t, err)

		schemaProto := sch.ToProto().(*schemapb.Schema) //nolint:forcetypeassert
		buildChan, buildChanCancel, err = client.build(ic.Context, connect.NewRequest(&langpb.BuildRequest{
			ProjectRoot: ic.WorkingDir(),
			StubsRoot:   filepath.Join(ic.WorkingDir(), ".ftl", config.Language, "modules"),
			BuildContext: &langpb.BuildContext{
				Id:           contextId,
				ModuleConfig: configProto,
				Schema:       schemaProto,
				Dependencies: dependencies,
			},
			RebuildAutomatically: rebuildAutomatically,
		}))
	}
}

func sendUpdatedBuildContext(contextId string, dependencies []string, sch *schema.Schema) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Sending updated context to plugin: %s", contextId)
		configProto, err := langpb.ModuleConfigToProto(config.Abs())
		assert.NoError(t, err)

		schemaProto := sch.ToProto().(*schemapb.Schema) //nolint:forcetypeassert
		_, err = client.buildContextUpdated(ic.Context, connect.NewRequest(&langpb.BuildContextUpdatedRequest{
			BuildContext: &langpb.BuildContext{
				Id:           contextId,
				ModuleConfig: configProto,
				Schema:       schemaProto,
				Dependencies: dependencies,
			},
		}))
		assert.NoError(t, err)
	}
}

func waitForAutoRebuildToStart(contextId string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Waiting for auto rebuild to start: %s", contextId)
		logger := log.FromContext(ic.Context)
		assert.NotZero(t, buildChan, "buildChan must be set before calling waitForAutoRebuildStarted")
		for {
			event, err := (<-buildChan).Result()
			assert.NoError(t, err, "did not expect a build stream error")
			switch event := event.Event.(type) {
			case *langpb.BuildEvent_AutoRebuildStarted:
				if event.AutoRebuildStarted.ContextId == contextId {
					return
				} else {
					logger.Warnf("ignoring automatic rebuild started event for unexpected context %q instead of %q", event.AutoRebuildStarted.ContextId, contextId)
				}
			case *langpb.BuildEvent_BuildSuccess:
				if event.BuildSuccess.ContextId == contextId {
					panic("build succeeded, but expected auto rebuild started event first")
				} else {
					logger.Warnf("ignoring build success for unexpected context %q while waiting for auto rebuild started event for %q", event.BuildSuccess.ContextId, contextId)
				}
			case *langpb.BuildEvent_BuildFailure:
				if event.BuildFailure.ContextId == contextId {
					panic("build failed, but expected auto rebuild started event first")
				} else {
					logger.Warnf("ignoring build failure for unexpected context %q while waiting for auto rebuild started event for %q", event.BuildFailure.ContextId, contextId)
				}
			}
		}
	}
}

func waitForBuildToEnd(success BuildResultType, contextId string, automaticRebuild bool, additionalChecks func(t testing.TB, ic in.TestContext, event *langpb.BuildEvent)) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		switch success {
		case SUCCESSORFAILURE:
			in.Infof("Waiting for build to end: %s", contextId)
		case SUCCESS:
			in.Infof("Waiting for build to succeed: %s", contextId)
		case FAILURE:
			in.Infof("Waiting for build to fail: %s", contextId)
		}
		logger := log.FromContext(ic.Context)
		assert.NotZero(t, buildChan, "buildChan must be set before calling waitForAutoRebuildStarted")
		for {
			e, err := (<-buildChan).Result()
			assert.NoError(t, err, "did not expect a build stream error")

			switch event := e.Event.(type) {
			case *langpb.BuildEvent_AutoRebuildStarted:
				if event.AutoRebuildStarted.ContextId != contextId {
					logger.Warnf("Ignoring automatic rebuild started event for unexpected context %q instead of %q", event.AutoRebuildStarted.ContextId, contextId)
					continue
				}
				logger.Debugf("Ignoring auto rebuild started event for the build we are waiting to finish %q", contextId)

			case *langpb.BuildEvent_BuildSuccess:
				if event.BuildSuccess.ContextId != contextId {
					logger.Warnf("Ignoring build success for unexpected context %q while waiting for auto rebuild started event for %q", event.BuildSuccess.ContextId, contextId)
					continue
				}
				if automaticRebuild != event.BuildSuccess.IsAutomaticRebuild {
					logger.Warnf("Ignoring build success for unexpected context %q (IsAutomaticRebuild=%v, expected=%v)", contextId, event.BuildSuccess.IsAutomaticRebuild, automaticRebuild)
					continue
				}
				if success == FAILURE {
					panic(fmt.Sprintf("build succeeded when we expected it to fail: %v", event.BuildSuccess))
				}
				if additionalChecks != nil {
					additionalChecks(t, ic, e)
				}
				return
			case *langpb.BuildEvent_BuildFailure:
				if event.BuildFailure.ContextId != contextId {
					logger.Warnf("Ignoring build failure for unexpected context %q while waiting for auto rebuild started event for %q", event.BuildFailure.ContextId, contextId)
					continue
				}
				if automaticRebuild != event.BuildFailure.IsAutomaticRebuild {
					logger.Warnf("Ignoring build failure for unexpected context %q (IsAutomaticRebuild=%v, expected=%v)", contextId, event.BuildFailure.IsAutomaticRebuild, automaticRebuild)
					continue
				}
				if success == SUCCESS {
					panic(fmt.Sprintf("build failed when we expected it to succeed: %v", event.BuildFailure))
				}
				if additionalChecks != nil {
					additionalChecks(t, ic, e)
				}
				return
			}
		}
	}
}

func checkForNoEvents(duration time.Duration) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Checking for no events for %v", duration)
		for {
			select {
			case result := <-buildChan:
				e, err := result.Result()
				assert.NoError(t, err, "did not expect a build stream error")
				switch event := e.Event.(type) {
				case *langpb.BuildEvent_AutoRebuildStarted:
					panic(fmt.Sprintf("rebuild started event when expecting no events: %v", event))
				case *langpb.BuildEvent_BuildSuccess:
					panic(fmt.Sprintf("build success event when expecting no events: %v", event))
				case *langpb.BuildEvent_BuildFailure:
					panic(fmt.Sprintf("build failure event when expecting no events: %v", event))
				}
			case <-time.After(duration):
				return
			case <-ic.Context.Done():
				return
			}
		}
	}
}

func addDependency(moduleName, depName string) in.Action {
	searchStr := "uncommentForDependency:"
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Adding dependency: %s", depName)
		found := false
		assert.NoError(t, walkWatchedFiles(t, ic, moduleName, func(path string) {
			bytes, err := os.ReadFile(path)
			assert.NoError(t, err)

			lines := strings.Split(string(bytes), "\n")

			foundInFile := false
			for i, line := range lines {
				start := strings.Index(line, searchStr)
				if start == -1 {
					continue
				}
				foundInFile = true
				end := start + len(searchStr)
				lines[i] = line[end:]
			}
			if foundInFile {
				found = true
				os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
			}
		}))
		assert.True(t, found, "could not add dependency because %q was not found in any files that are watched", searchStr)
	}
}

func modifyVerbName(moduleName, old, new string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Modifying verb name: %s -> %s", old, new)
		found := false
		assert.NoError(t, walkWatchedFiles(t, ic, moduleName, func(path string) {
			bytes, err := os.ReadFile(path)
			assert.NoError(t, err)

			lines := strings.Split(string(bytes), "\n")

			foundInFile := false
			for i, line := range lines {
				start := strings.Index(line, old)
				if start == -1 {
					continue
				}
				foundInFile = true
				end := start + len(old)
				lines[i] = line[:start] + new + line[end:]
			}
			if foundInFile {
				found = true
				os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
			}
		}))
		assert.True(t, found, "could not modify verb name because %q was not found in any files that are watched", old)
	}
}

func walkWatchedFiles(t testing.TB, ic in.TestContext, moduleName string, visit func(path string)) error {
	path := filepath.Join(ic.WorkingDir(), moduleName)
	return watch.WalkDir(path, func(srcPath string, entry fs.DirEntry) error {
		if entry.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(path, srcPath)
		assert.NoError(t, err)

		_, matched := slices.Find(config.Watch, func(pattern string) bool {
			match, err := doublestar.PathMatch(pattern, relativePath)
			assert.NoError(t, err)
			return match
		})
		if matched {
			visit(srcPath)
		}
		return nil
	})
}

func checkBuildLockLifecycle(childActions ...in.Action) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Checking build lock is unlocked")
		_, err := os.Stat(config.Abs().BuildLock)
		assert.Error(t, err, "expected build lock file to not exist before building at %v", config.Abs().BuildLock)

		lockFound := make(chan bool)
		go func() {
			defer close(lockFound)
			startTime := time.Now()
			for {
				select {
				case <-time.After(1 * time.Second):
					if _, err := os.Stat(config.Abs().BuildLock); err == nil {
						lockFound <- true
						return
					}
					if time.Since(startTime) > 3*time.Second {
						lockFound <- false
						return
					}
				case <-ic.Context.Done():
					lockFound <- false
					return
				}
			}
		}()

		// do build actions
		for _, childAction := range childActions {
			childAction(t, ic)
		}
		// confirm that at some point we did find the lock file
		assert.True(t, (<-lockFound), "never found build lock file at %v while building", config.Abs().BuildLock)

		in.Infof("Checking build lock is unlocked")
		_, err = os.Stat(config.Abs().BuildLock)
		assert.Error(t, err, "expected build lock file to not exist after building at %v", config.Abs().BuildLock)
	}
}

func obtainAndReleaseBuildLock(duration time.Duration) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Obtaining and releasing build lock")
		release, err := flock.Acquire(ic.Context, config.Abs().BuildLock, BuildLockTimeout)
		assert.NoError(t, err, "could not get build lock")
		time.Sleep(duration)
		err = release()
		assert.NoError(t, err, "could not release build lock")
	}
}

func checkLockedBehavior(buildFailureActions ...in.Action) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Acquiring build lock: %v", config.Abs().BuildLock)
		release, err := flock.Acquire(ic.Context, config.Abs().BuildLock, BuildLockTimeout)
		assert.NoError(t, err, "could not get build lock")

		// build on a separate goroutine
		buildEnded := make(chan bool)
		go func() {
			for _, buildAction := range buildFailureActions {
				buildAction(t, ic)
			}
			close(buildEnded)
		}()

		// wait for build to fail due to file lock
		_ = <-buildEnded

		err = release() //nolint:errcheck
		assert.NoError(t, err, "could not release build lock")
	}
}
