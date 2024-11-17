//go:build integration || infrastructure || smoketest

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver
	"github.com/kballard/go-shellquote"
	"github.com/otiai10/copy"
	"k8s.io/client-go/kubernetes"

	"github.com/block/scaffolder"

	"github.com/TBD54566975/ftl/backend/controller/dsn"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	ftlexec "github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

// Scaffold a directory relative to the testdata directory to a directory relative to the working directory.
func Scaffold(src, dest string, tmplCtx any) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Scaffolding %s -> %s", src, dest)
		err := scaffolder.Scaffold(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest), tmplCtx)
		assert.NoError(t, err)
	}
}

// GitInit calls git init on the working directory.
func GitInit() Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Running `git init` on the working directory: %s", ic.workDir)
		err := ftlexec.Command(ic, log.Debug, ic.workDir, "git", "init", ic.workDir).RunBuffered(ic)
		assert.NoError(t, err)
	}
}

// Copy a module from the testdata directory to the working directory.
//
// Ensures that any language-specific local modifications are made correctly,
// such as Go module file replace directives for FTL.
func CopyModule(module string) Action {
	return Chain(
		CopyDir(module, module),
		editGoMod(module),
	)
}

// Copy a module from the testdata directory to the working directory
// This will always use the specified language regardless of the language under test
//
// Ensures that any language-specific local modifications are made correctly,
// such as Go module file replace directives for FTL.
func CopyModuleWithLanguage(module string, language string) Action {
	return Chain(
		CopyDir(filepath.Join("..", language, module), module),
		editGoMod(module),
	)
}

func editGoMod(module string) func(t testing.TB, ic TestContext) {
	return func(t testing.TB, ic TestContext) {
		root := filepath.Join(ic.workDir, module)
		// TODO: Load the module configuration from the module itself and use that to determine the language-specific stuff.
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			err := ftlexec.Command(ic, log.Debug, root, "go", "mod", "edit", "-replace", "github.com/TBD54566975/ftl="+ic.RootDir).RunBuffered(ic)
			assert.NoError(t, err)
		}
	}
}

// SetEnv sets an environment variable for the duration of the test.
//
// Note that the FTL controller will already be running.
func SetEnv(key string, value func(ic TestContext) string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Setting environment variable %s=%s", key, value(ic))
		t.Setenv(key, value(ic))
	}
}

// Copy a directory from the testdata directory to the working directory.
func CopyDir(src, dest string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Copying %s -> %s", src, dest)
		err := copy.Copy(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest))
		assert.NoError(t, err)
	}
}

// Chain multiple actions together.
func Chain(actions ...Action) Action {
	return func(t testing.TB, ic TestContext) {
		for _, action := range actions {
			action(t, ic)
		}
	}
}

// SubTests runs a list of individual actions as separate tests
func SubTests(tests ...SubTest) Action {
	return func(t testing.TB, ic TestContext) {
		for _, test := range tests {
			ic.Run(test.Name, func(t *testing.T) {
				ic.AssertWithRetry(t, test.Action)
			})
		}
	}
}

// Repeat an action N times.
func Repeat(n int, action Action) Action {
	return func(t testing.TB, ic TestContext) {
		for i := 0; i < n; i++ {
			action(t, ic)
		}
	}
}

// Chdir changes the test working directory to the subdirectory for the duration of the action.
func Chdir(dir string, a Action) Action {
	return func(t testing.TB, ic TestContext) {
		dir := filepath.Join(ic.workDir, dir)
		Infof("Changing directory to %s", dir)
		cwd, err := os.Getwd()
		assert.NoError(t, err)
		ic.workDir = dir
		err = os.Chdir(dir)
		assert.NoError(t, err)
		defer os.Chdir(cwd)
		a(t, ic)
	}
}

// DebugShell opens a new Terminal window in the test working directory.
func DebugShell() Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Starting debug shell")
		err := ftlexec.Command(ic, log.Debug, ic.workDir, "open", "-n", "-W", "-a", "Terminal", ".").RunBuffered(ic)
		assert.NoError(t, err)
	}
}

// Exec runs a command from the test working directory.
func Exec(cmd string, args ...string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Executing (in %s): %s %s", ic.workDir, cmd, shellquote.Join(args...))
		err := ftlexec.Command(ic, log.Debug, ic.workDir, cmd, args...).RunStderrError(ic)
		assert.NoError(t, err)
	}
}

// ExecWithExpectedOutput runs a command from the test working directory.
// The output is captured and is compared with the expected output.
func ExecWithExpectedOutput(want string, cmd string, args ...string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Executing: %s %s", cmd, shellquote.Join(args...))
		output, err := ftlexec.Capture(ic, ic.workDir, cmd, args...)
		assert.NoError(t, err)
		assert.Equal(t, output, []byte(want))
	}
}

// ExecWithExpectedError runs a command from the test working directory, and
// expects it to fail with the given error message.
func ExecWithExpectedError(want string, cmd string, args ...string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Executing: %s %s", cmd, shellquote.Join(args...))
		output, err := ftlexec.Capture(ic, ic.workDir, cmd, args...)
		assert.Error(t, err)
		assert.Contains(t, string(output), want)
	}
}

// ExecWithOutput runs a command from the test working directory.
// On success capture() is executed with the output
// On error, an error with the output is returned.
func ExecWithOutput(cmd string, args []string, capture func(output string)) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Executing: %s %s", cmd, shellquote.Join(args...))
		output, err := ftlexec.Capture(ic, ic.workDir, cmd, args...)
		assert.NoError(t, err, "%s", string(output))
		capture(string(output))
	}
}

// ExpectError wraps an action and expects it to return an error containing the given messages.
func ExpectError(action Action, expectedErrorMsg ...string) Action {
	return func(t testing.TB, ic TestContext) {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(TestingError); ok {
					for _, msg := range expectedErrorMsg {
						assert.Contains(t, string(e), msg)
					}
				} else {
					panic(r)
				}
			}
		}()
		action(t, ic)
	}
}

// Deploy a module from the working directory and wait for it to become available.
func Deploy(module string) Action {
	return Chain(
		func(t testing.TB, ic TestContext) {
			args := []string{"deploy"}
			if ic.Provisioner != nil {
				args = append(args, "--use-provisioner", "--provisioner-endpoint=http://localhost:8893")
			}
			if ic.kubeClient.Ok() {
				args = append(args, "--build-env", "GOOS=linux", "--build-env", "GOARCH=amd64", "--build-env", "CGO_ENABLED=0", "--use-provisioner")
			}
			args = append(args, module)

			Exec("ftl", args...)(t, ic)
		},
		Wait(module),
	)
}

// Build modules from the working directory and wait for it to become available.
func Build(modules ...string) Action {
	args := []string{"build"}
	args = append(args, modules...)
	return Exec("ftl", args...)
}

// FtlNew creates a new FTL module
func FtlNew(language, name string) Action {
	return func(t testing.TB, ic TestContext) {
		err := ftlexec.Command(ic, log.Debug, ic.workDir, "ftl", "new", language, ic.workDir, name).RunBuffered(ic)
		assert.NoError(t, err)
	}
}

// Wait for the given module to deploy.
func Wait(module string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Waiting for %s to become ready", module)
		// There's a bit of a bug here: wait() is already being retried by the
		// test harness, so in the error case we'll be waiting N^2 times. This
		// is fine for now, but we should fix this in the future.
		ic.AssertWithRetry(t, func(t testing.TB, ic TestContext) {
			status, err := ic.Controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
			assert.NoError(t, err)
			for _, deployment := range status.Msg.Deployments {
				if deployment.Name == module {
					return
				}
			}
			t.Fatalf("deployment of module %q not found", module)
		})
	}
}

// WaitWithTimeout for the given module to deploy.
func WaitWithTimeout(module string, timeout time.Duration) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Waiting for %s to become ready", module)
		deadline := time.After(timeout)
		tick := time.NewTicker(time.Millisecond * 100)
		defer tick.Stop()
		for {
			select {
			case <-deadline:
				t.Fatalf("deployment of module %q not found", module)
				return
			case <-tick.C:
				status, err := ic.Controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
				assert.NoError(t, err)
				for _, deployment := range status.Msg.Deployments {
					if deployment.Name == module {
						return
					}
				}
			}
		}
	}
}

func Sleep(duration time.Duration) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Sleeping for %s", duration)
		time.Sleep(duration)
	}
}

// Assert that a file exists in the working directory.
func FileExists(path string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Checking that %s exists", path)
		_, err := os.Stat(filepath.Join(ic.workDir, path))
		assert.NoError(t, err)
	}
}

// Assert that a file exists and its content contains the given text.
//
// If "path" is relative it will be to the working directory.
func FileContains(path, needle string) Action {
	return func(t testing.TB, ic TestContext) {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		Infof("Checking that the content of %s is correct", absPath)
		data, err := os.ReadFile(absPath)
		assert.NoError(t, err)
		actual := string(data)
		assert.Contains(t, actual, needle)
	}
}

// Assert that a file exists and its content is equal to the given text.
//
// If "path" is relative it will be to the working directory.
func FileContent(path, expected string) Action {
	return func(t testing.TB, ic TestContext) {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		Infof("Checking that the content of %s is correct", absPath)
		data, err := os.ReadFile(absPath)
		assert.NoError(t, err)
		expected = strings.TrimSpace(expected)
		actual := strings.TrimSpace(string(data))
		assert.Equal(t, expected, actual)
	}
}

// WriteFile writes a file to the working directory.
func WriteFile(path string, content []byte) Action {
	return func(t testing.TB, ic TestContext) {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		Infof("Writing to %s", path)
		err := os.WriteFile(absPath, content, 0600)
		assert.NoError(t, err)
	}
}

// EditFile edits a file in a module
func EditFile(module string, editFunc func([]byte) []byte, path ...string) Action {
	return func(t testing.TB, ic TestContext) {
		parts := []string{ic.workDir, module}
		parts = append(parts, path...)
		file := filepath.Join(parts...)
		Infof("Editing %s", file)
		contents, err := os.ReadFile(file)
		assert.NoError(t, err)
		contents = editFunc(contents)
		err = os.WriteFile(file, contents, os.FileMode(0))
		assert.NoError(t, err)
	}
}

// MoveFile moves a file within a module
func MoveFile(module, from, to string) Action {
	return func(t testing.TB, ic TestContext) {
		err := os.Rename(filepath.Join(ic.WorkingDir(), module, from), filepath.Join(ic.WorkingDir(), module, to))
		assert.NoError(t, err)
	}
}

// RemoveDir removes the given directory and all of its contents under the working dir
func RemoveDir(dir string) Action {
	return func(t testing.TB, ic TestContext) {
		err := os.RemoveAll(filepath.Join(ic.WorkingDir(), dir))
		assert.NoError(t, err)
	}
}

type Obj map[string]any

// Call a verb.
//
// "check" may be nil
func Call[Req any, Resp any](module, verb string, request Req, check func(t testing.TB, response Resp)) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Calling %s.%s", module, verb)
		assert.False(t, unicode.IsUpper([]rune(verb)[0]), "verb %q must start with an lowercase letter", verb)
		data, err := json.Marshal(request)
		assert.NoError(t, err)
		resp, err := ic.Verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
			Verb: &schemapb.Ref{Module: module, Name: verb},
			Body: data,
		}))
		assert.NoError(t, err)
		var response Resp
		assert.Zero(t, resp.Msg.GetError(), "verb failed: %s", resp.Msg.GetError().GetMessage())
		err = json.Unmarshal(resp.Msg.GetBody(), &response)
		assert.NoError(t, err)
		if check != nil {
			check(t, response)
		}
	}
}

// VerifyKubeState lets you test the current kube state
func VerifyKubeState(check func(ctx context.Context, t testing.TB, namespace string, client kubernetes.Clientset)) Action {
	return func(t testing.TB, ic TestContext) {
		check(ic.Context, t, ic.kubeNamespace, ic.kubeClient.MustGet())
	}
}

// VerifySchema lets you test the current schema
func VerifySchema(check func(ctx context.Context, t testing.TB, sch *schemapb.Schema)) Action {
	return func(t testing.TB, ic TestContext) {
		sch, err := ic.Schema.GetSchema(ic, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
		if err != nil {
			t.Errorf("failed to get schema: %v", err)
			return
		}
		check(ic.Context, t, sch.Msg.GetSchema())
	}
}

// VerifyControllerStatus lets you test the current controller status
func VerifyControllerStatus(check func(ctx context.Context, t testing.TB, status *ftlv1.StatusResponse)) Action {
	return func(t testing.TB, ic TestContext) {
		sch, err := ic.Controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
		if err != nil {
			t.Errorf("failed to get schema: %v", err)
			return
		}
		check(ic.Context, t, sch.Msg)
	}
}

// VerifySchemaVerb lets you test the current schema for a specific verb
func VerifySchemaVerb(module string, verb string, check func(ctx context.Context, t testing.TB, schema *schemapb.Schema, verb *schemapb.Verb)) Action {
	return func(t testing.TB, ic TestContext) {
		sch, err := ic.Schema.GetSchema(ic, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
		if err != nil {
			t.Errorf("failed to get schema: %v", err)
			return
		}
		for _, m := range sch.Msg.GetSchema().Modules {
			if m.Name == module {
				for _, v := range m.Decls {
					if v.GetVerb() != nil && v.GetVerb().Name == verb {
						check(ic.Context, t, sch.Msg.GetSchema(), v.GetVerb())
						return
					}
				}
			}
		}
		t.Errorf("verb %s.%s not found in schema", module, verb)
	}
}

// Fail expects the next action to Fail.
func Fail(next Action, msg string, args ...any) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Expecting failure of nested action")
		panicked := true
		defer func() {
			if !panicked {
				t.Fatalf("expected action to fail: "+msg, args...)
			} else {
				recover()
			}
		}()
		next(t, ic)
		panicked = false
	}
}

// fetched and returns a row's column values
func GetRow(t testing.TB, ic TestContext, database, query string, fieldCount int) []any {
	Infof("Querying %s: %s", database, query)
	db, err := sql.Open("pgx", dsn.DSN(database))
	assert.NoError(t, err)
	defer db.Close()
	actual := make([]any, fieldCount)
	for i := range actual {
		actual[i] = new(any)
	}
	err = db.QueryRowContext(ic, query).Scan(actual...)
	assert.NoError(t, err)
	for i := range actual {
		actual[i] = *actual[i].(*any)
	}
	return actual
}

// Query a single row from a database.
func QueryRow(database string, query string, expected ...interface{}) Action {
	return func(t testing.TB, ic TestContext) {
		actual := GetRow(t, ic, database, query, len(expected))
		for i, a := range actual {
			assert.Equal(t, expected[i], a)
		}
	}
}

// Create a database for use by a module.
func CreateDBAction(module, dbName string, isTest bool) Action {
	return func(t testing.TB, ic TestContext) {
		CreateDB(t, module, dbName, isTest)
	}
}

func terminateDanglingConnections(t testing.TB, db *sql.DB, dbName string) {
	t.Helper()

	_, err := db.Exec(`
		SELECT pid, pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()`,
		dbName)
	assert.NoError(t, err)
}

func CreateDB(t testing.TB, module, dbName string, isTestDb bool) {
	// insert test suffix if needed when actually setting up db
	if isTestDb {
		dbName += "_test"
	}
	Infof("Creating database %s", dbName)
	db, err := sql.Open("pgx", dsn.DSN("ftl"))
	assert.NoError(t, err, "failed to open database connection")
	t.Cleanup(func() {
		err := db.Close()
		assert.NoError(t, err)
	})

	err = db.Ping()
	assert.NoError(t, err, "failed to ping database")

	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	assert.NoError(t, err, "failed to delete existing database")

	_, err = db.Exec("CREATE DATABASE " + dbName)
	assert.NoError(t, err, "failed to create database")

	t.Cleanup(func() {
		terminateDanglingConnections(t, db, dbName)
		_, err = db.Exec("DROP DATABASE " + dbName)
		assert.NoError(t, err)
	})
}

func DropDBAction(t testing.TB, dbName string) Action {
	return func(t testing.TB, ic TestContext) {
		DropDB(t, dbName)
	}
}

func DropDB(t testing.TB, dbName string) {
	Infof("Dropping database %s", dbName)
	db, err := sql.Open("pgx", dsn.DSN("postgres"))
	assert.NoError(t, err, "failed to open database connection")

	terminateDanglingConnections(t, db, dbName)

	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	assert.NoError(t, err, "failed to delete existing database")

	t.Cleanup(func() {
		err := db.Close()
		assert.NoError(t, err)
	})
}

// Create a directory in the working directory
func Mkdir(dir string) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("Creating directory %s", dir)
		err := os.MkdirAll(filepath.Join(ic.workDir, dir), 0700)
		assert.NoError(t, err)
	}
}

type HTTPResponse struct {
	Status    int
	Headers   map[string][]string
	JsonBody  map[string]any
	BodyBytes []byte
}

func JsonData(t testing.TB, body interface{}) []byte {
	b, err := json.Marshal(body)
	assert.NoError(t, err)
	return b
}

// HttpCall makes an HTTP call to the running FTL ingress endpoint.
func HttpCall(method string, path string, headers map[string][]string, body []byte, onResponse func(t testing.TB, resp *HTTPResponse)) Action {
	return func(t testing.TB, ic TestContext) {
		Infof("HTTP %s %s", method, path)
		baseURL, err := url.Parse(fmt.Sprintf("http://localhost:8891"))
		assert.NoError(t, err)

		u, err := baseURL.Parse(path)
		assert.NoError(t, err)
		r, err := http.NewRequestWithContext(ic, method, u.String(), bytes.NewReader(body))
		assert.NoError(t, err)

		r.Header.Add("Content-Type", "application/json")
		for k, vs := range headers {
			for _, v := range vs {
				r.Header.Add(k, v)
			}
		}

		client := http.Client{}
		resp, err := client.Do(r)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var resBody map[string]any
		// ignore the error here since some responses are just `[]byte`.
		_ = json.Unmarshal(bodyBytes, &resBody)

		onResponse(t, &HTTPResponse{
			Status:    resp.StatusCode,
			Headers:   resp.Header,
			JsonBody:  resBody,
			BodyBytes: bodyBytes,
		})
	}
}

func IfLanguage(language string, action Action) Action {
	return IfLanguages(action, language)
}

func IfLanguages(action Action, languages ...string) Action {
	return func(t testing.TB, ic TestContext) {
		if slices.Contains(languages, ic.Language) {
			action(t, ic)
		}
	}
}

// Run "go test" in the given module.
func ExecModuleTest(module string) Action {
	return Chdir(module, Exec("go", "test", "./..."))
}
