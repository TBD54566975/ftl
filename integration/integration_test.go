//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/iancoleman/strcase"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/rpc"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/scaffolder"
)

const integrationTestTimeout = time.Second * 60

var runtimes = []string{"go", "kotlin"}

func TestLifecycle(t *testing.T) {
	runForRuntimes(t, func(modulesDir string, runtime string, rd runtimeData) []test {
		return []test{
			{name: fmt.Sprintf("Init%s", rd.testSuffix), assertions: assertions{
				run("ftl", rd.initOpts...),
			}},
			{name: fmt.Sprintf("Deploy%s", rd.testSuffix), assertions: assertions{
				run("ftl", "deploy", rd.moduleRoot),
				deploymentExists(rd.moduleName),
			}},
			{name: fmt.Sprintf("Call%s", rd.testSuffix), assertions: assertions{
				call(rd.moduleName, "echo", obj{"name": "Alice"}, func(t testing.TB, resp obj) {
					message, ok := resp["message"].(string)
					assert.True(t, ok, "message is not a string")
					assert.True(t, regexp.MustCompile(`^Hello, Alice!`).MatchString(message), "%q does not match %q", message, `^Hello, Alice!`)
				}),
			}},
		}
	})
}
func TestHttpIngress(t *testing.T) {
	runForRuntimes(t, func(modulesDir string, runtime string, rd runtimeData) []test {
		return []test{
			{name: fmt.Sprintf("HttpIngress%s", rd.testSuffix), assertions: assertions{
				run("ftl", rd.initOpts...),
				scaffoldTestData(runtime, "httpingress", rd.modulePath),
				run("ftl", "deploy", rd.moduleRoot),
				httpCall(rd, http.MethodGet, "/users/123/posts/456", jsonData(t, obj{}), func(t testing.TB, resp *httpResponse) {
					assert.Equal(t, 200, resp.status)
					assert.Equal(t, []string{"Header from FTL"}, resp.headers["Get"])
					assert.Equal(t, []string{"application/json; charset=utf-8"}, resp.headers["Content-Type"])

					message, ok := resp.body["msg"].(string)
					assert.True(t, ok, "msg is not a string")
					assert.Equal(t, "UserID: 123, PostID: 456", message)

					nested, ok := resp.body["nested"].(map[string]any)
					assert.True(t, ok, "nested is not a map")
					goodStuff, ok := nested["good_stuff"].(string)
					assert.True(t, ok, "good_stuff is not a string")
					assert.Equal(t, "This is good stuff", goodStuff)
				}),
				httpCall(rd, http.MethodPost, "/users", jsonData(t, obj{"userID": 123, "postID": 345}), func(t testing.TB, resp *httpResponse) {
					assert.Equal(t, 201, resp.status)
					assert.Equal(t, []string{"Header from FTL"}, resp.headers["Post"])
					success, ok := resp.body["success"].(bool)
					assert.True(t, ok, "success is not a bool")
					assert.True(t, success)
				}),
				// contains aliased field
				httpCall(rd, http.MethodPost, "/users", jsonData(t, obj{"user_id": 123, "postID": 345}), func(t testing.TB, resp *httpResponse) {
					assert.Equal(t, 201, resp.status)
				}),
				httpCall(rd, http.MethodPut, "/users/123", jsonData(t, obj{"postID": "346"}), func(t testing.TB, resp *httpResponse) {
					assert.Equal(t, 200, resp.status)
					assert.Equal(t, []string{"Header from FTL"}, resp.headers["Put"])
					assert.Equal(t, nil, resp.body)
				}),
				httpCall(rd, http.MethodDelete, "/users/123", jsonData(t, obj{}), func(t testing.TB, resp *httpResponse) {
					assert.Equal(t, 200, resp.status)
					assert.Equal(t, []string{"Header from FTL"}, resp.headers["Delete"])
					assert.Equal(t, nil, resp.body)
				}),

				httpCall(rd, http.MethodGet, "/html", jsonData(t, obj{}), func(t testing.TB, resp *httpResponse) {
					assert.Equal(t, 200, resp.status)
					assert.Equal(t, []string{"text/html; charset=utf-8"}, resp.headers["Content-Type"])
					assert.Equal(t, "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>", string(resp.bodyBytes))
				}),

				httpCall(rd, http.MethodPost, "/bytes", []byte("Hello, World!"), func(t testing.TB, resp *httpResponse) {
					assert.Equal(t, 200, resp.status)
					assert.Equal(t, []string{"application/octet-stream"}, resp.headers["Content-Type"])
					assert.Equal(t, []byte("Hello, World!"), resp.bodyBytes)
				}),

				httpCall(rd, http.MethodGet, "/empty", nil, func(t testing.TB, resp *httpResponse) {
					assert.Equal(t, 200, resp.status)
					assert.Equal(t, nil, resp.headers["Content-Type"])
					assert.Equal(t, nil, resp.bodyBytes)
				}),

				httpCall(rd, http.MethodGet, "/string", []byte("Hello, World!"), func(t testing.TB, resp *httpResponse) {
					assert.Equal(t, 200, resp.status)
					assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.headers["Content-Type"])
					assert.Equal(t, []byte("Hello, World!"), resp.bodyBytes)
				}),
			}},
		}
	})
}

func TestDatabase(t *testing.T) {
	runForRuntimes(t, func(modulesDir string, runtime string, rd runtimeData) []test {
		dbName := "testdb"
		err := os.Setenv(
			fmt.Sprintf("FTL_POSTGRES_DSN_%s_TESTDB", strings.ToUpper(rd.moduleName)),
			fmt.Sprintf("postgres://postgres:secret@localhost:54320/%s?sslmode=disable", dbName),
		)
		assert.NoError(t, err)

		requestData := fmt.Sprintf("Hello %s", runtime)
		return []test{
			{name: fmt.Sprintf("Database%s", rd.testSuffix), assertions: assertions{
				setUpModuleDB(dbName),
				run("ftl", rd.initOpts...),
				scaffoldTestData(runtime, "database", rd.modulePath),
				run("ftl", "deploy", rd.moduleRoot),
				call(rd.moduleName, "insert", obj{"data": requestData}, func(t testing.TB, resp obj) {}),
				validateModuleDB(dbName, requestData),
			}},
		}
	})
}

func TestExternalCalls(t *testing.T) {
	runForRuntimes(t, func(modulesDir string, runtime string, rd runtimeData) []test {
		var tests []test
		for _, callee := range runtimes {
			calleeRd := getRuntimeData("echo2", modulesDir, callee)
			tests = append(tests, test{
				name: fmt.Sprintf("Call%sFrom%s", strcase.ToCamel(callee), strcase.ToCamel(runtime)),
				assertions: assertions{
					run("ftl", calleeRd.initOpts...),
					run("ftl", "deploy", calleeRd.moduleRoot),
					run("ftl", rd.initOpts...),
					scaffoldTestData(runtime, "externalcalls", rd.modulePath),
					run("ftl", "deploy", rd.moduleRoot),
					call(rd.moduleName, "call", obj{"name": "Alice"}, func(t testing.TB, resp obj) {
						message, ok := resp["message"].(string)
						assert.True(t, ok, "message is not a string")
						assert.True(t, regexp.MustCompile(`^Hello, Alice!`).MatchString(message), "%q does not match %q", message, `^Hello, Alice!`)
					}),
					run("rm", "-rf", rd.moduleRoot),
					run("rm", "-rf", calleeRd.moduleRoot),
				}})
		}
		return tests
	})
}

func TestSchemaGenerate(t *testing.T) {
	tests := []test{
		{name: "SchemaGenerateJS", assertions: assertions{
			run("ftl", "schema", "generate", "integration/testdata/schema-generate", "build/schema-generate"),
			filesExist(file{"build/schema-generate/test.txt", "olleh"}),
		}},
	}
	runTests(t, t.TempDir(), tests)
}

type testsFunc func(modulesDir string, runtime string, rd runtimeData) []test

func runForRuntimes(t *testing.T, f testsFunc) {
	t.Helper()
	tmpDir := t.TempDir()
	modulesDir := filepath.Join(tmpDir, "modules")

	var tests []test
	for _, runtime := range runtimes {
		rd := getRuntimeData("echo", modulesDir, runtime)
		tests = append(tests, f(modulesDir, runtime, rd)...)
	}
	runTests(t, tmpDir, tests)
}

type test struct {
	name       string
	assertions assertions
}

func runTests(t *testing.T, tmpDir string, tests []test) {
	t.Helper()
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	rootDir := filepath.Join(cwd, "..")

	// Build FTL binary
	logger := log.Configure(&logWriter{logger: t}, log.Config{Level: log.Debug})
	ctx := log.ContextWithLogger(context.Background(), logger)
	logger.Infof("Building ftl")
	binDir := filepath.Join(rootDir, "build", "release")
	err = exec.Command(ctx, log.Debug, rootDir, filepath.Join(rootDir, "bin", "bit"), "build/release/ftl", "**/*.jar").Run()
	assert.NoError(t, err)

	controller := rpc.Dial(ftlv1connect.NewControllerServiceClient, "http://localhost:8892", log.Debug)
	verbs := rpc.Dial(ftlv1connect.NewVerbServiceClient, "http://localhost:8892", log.Debug)

	ctx = startProcess(ctx, t, filepath.Join(binDir, "ftl"), "serve", "--recreate")

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

type runtimeData struct {
	testSuffix string
	moduleName string
	moduleRoot string
	modulePath string
	initOpts   []string
}

func getRuntimeData(moduleName string, modulesDir string, runtime string) runtimeData {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	ftlRoot := filepath.Join(cwd, "..")

	t := runtimeData{
		testSuffix: strcase.ToCamel(runtime),
		moduleName: moduleName,
	}
	switch runtime {
	case "go":
		t.moduleRoot = filepath.Join(modulesDir, t.moduleName)
		t.modulePath = t.moduleRoot
		// include replace flag to use local ftl in go.mod
		t.initOpts = []string{"init", runtime, modulesDir, t.moduleName, "--replace", fmt.Sprintf("github.com/TBD54566975/ftl=%s", ftlRoot)}
	case "kotlin":
		t.moduleRoot = filepath.Join(modulesDir, fmt.Sprintf("ftl-module-%s", t.moduleName))
		t.modulePath = filepath.Join(t.moduleRoot, "src/main/kotlin/ftl", t.moduleName)
		t.initOpts = []string{"init", runtime, modulesDir, t.moduleName}
	default:
		panic(fmt.Sprintf("unknown runtime %q", runtime))
	}
	return t
}

type assertion func(t testing.TB, ic itContext) error
type assertions []assertion

// Assertions

// Run a command in "dir" which is relative to the root directory of the project.
func run(cmd string, args ...string) assertion {
	return func(t testing.TB, ic itContext) error {
		path := os.Getenv("PATH")
		path = ic.binDir + ":" + path
		cmd := exec.Command(ic, log.Debug, ic.rootDir, cmd, args...)
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env, "PATH="+path)
		return cmd.Run()
	}
}

func deploymentExists(name string) assertion {
	return status(func(t testing.TB, status *ftlv1.StatusResponse) {
		for _, deployment := range status.Deployments {
			if deployment.Schema.Name == name {
				return
			}
		}
		t.Fatal(fmt.Sprintf("%s deployment not found", name))
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

type httpResponse struct {
	status    int
	headers   map[string][]string
	body      map[string]any
	bodyBytes []byte
}

func jsonData(t testing.TB, body obj) []byte {
	b, err := json.Marshal(body)
	assert.NoError(t, err)
	return b
}

func httpCall(rd runtimeData, method string, path string, body []byte, onResponse func(t testing.TB, resp *httpResponse)) assertion {
	return func(t testing.TB, ic itContext) error {
		baseURL, err := url.Parse(fmt.Sprintf("http://localhost:8892/ingress"))
		assert.NoError(t, err)

		r, err := http.NewRequestWithContext(ic, method, baseURL.JoinPath(path).String(), bytes.NewReader(body))
		assert.NoError(t, err)

		r.Header.Add("Content-Type", "application/json")

		client := http.Client{}
		resp, err := client.Do(r)
		assert.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			// error here so that the test retries in case the 404 is caused by the runner not being ready
			return fmt.Errorf("endpoint not found: %s", path)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var resBody map[string]any
		// ignore the error here since some responses are just `[]byte`.
		_ = json.Unmarshal(bodyBytes, &resBody)

		onResponse(t, &httpResponse{
			status:    resp.StatusCode,
			headers:   resp.Header,
			body:      resBody,
			bodyBytes: bodyBytes,
		})
		return nil
	}
}

func scaffoldTestData(runtime string, testDataDirectory string, targetModulePath string) assertion {
	return func(t testing.TB, ic itContext) error {
		return scaffolder.Scaffold(
			filepath.Join(ic.rootDir, fmt.Sprintf("integration/testdata/%s/", runtime), testDataDirectory),
			targetModulePath,
			ic,
		)
	}
}

func setUpModuleDB(dbName string) assertion {
	return func(t testing.TB, ic itContext) error {
		db, err := sql.Open("pgx", "postgres://postgres:secret@localhost:54320/ftl?sslmode=disable")
		assert.NoError(t, err)
		t.Cleanup(func() {
			err := db.Close()
			if err != nil {
				t.Fatal(err)
			}
		})

		err = db.Ping()
		assert.NoError(t, err)

		var exists bool
		query := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1);`
		err = db.QueryRow(query, dbName).Scan(&exists)
		assert.NoError(t, err)
		if !exists {
			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
			assert.NoError(t, err)
		}

		return nil
	}
}
func validateModuleDB(dbName string, expectedRowData string) assertion {
	return func(t testing.TB, ic itContext) error {
		db, err := sql.Open("pgx", fmt.Sprintf("postgres://postgres:secret@localhost:54320/%s?sslmode=disable", dbName))
		assert.NoError(t, err)
		t.Cleanup(func() {
			err := db.Close()
			if err != nil {
				t.Fatal(err)
			}
		})

		err = db.Ping()
		assert.NoError(t, err)

		rows, err := db.Query("SELECT data FROM requests")
		defer rows.Close()
		assert.NoError(t, err)

		for rows.Next() {
			var data string
			err := rows.Scan(&data)
			assert.NoError(t, err)
			if data == expectedRowData {
				return nil
			}
		}

		return errors.New("data not found")
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
func startProcess(ctx context.Context, t testing.TB, args ...string) context.Context {
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
		}
		l.buffer = append(l.buffer, p[:index]...)
		l.logger.Log(string(l.buffer))
		l.buffer = l.buffer[:0]
		p = p[index+1:]
	}
}
