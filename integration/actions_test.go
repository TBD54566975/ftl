//go:build integration

package simple_test

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
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	gensql "github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver
	"github.com/kballard/go-shellquote"
	"github.com/otiai10/copy"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlexec "github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/scaffolder"
)

// scaffold a directory relative to the testdata directory to a directory relative to the working directory.
func scaffold(src, dest string, tmplCtx any) action {
	return func(t testing.TB, ic testContext) {
		infof("Scaffolding %s -> %s", src, dest)
		err := scaffolder.Scaffold(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest), tmplCtx)
		assert.NoError(t, err)
	}
}

// Copy a module from the testdata directory to the working directory.
//
// Ensures that replace directives are correctly handled.
func copyModule(module string) action {
	return chain(
		copyDir(module, module),
		func(t testing.TB, ic testContext) {
			err := ftlexec.Command(ic, log.Debug, filepath.Join(ic.workDir, module), "go", "mod", "edit", "-replace", "github.com/TBD54566975/ftl="+ic.rootDir).RunBuffered(ic)
			assert.NoError(t, err)
		},
	)
}

// Copy a directory from the testdata directory to the working directory.
func copyDir(src, dest string) action {
	return func(t testing.TB, ic testContext) {
		infof("Copying %s -> %s", src, dest)
		err := copy.Copy(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest))
		assert.NoError(t, err)
	}
}

// chain multiple actions together.
func chain(actions ...action) action {
	return func(t testing.TB, ic testContext) {
		for _, action := range actions {
			action(t, ic)
		}
	}
}

// chdir changes the test working directory to the subdirectory for the duration of the action.
func chdir(dir string, a action) action {
	return func(t testing.TB, ic testContext) {
		dir := filepath.Join(ic.workDir, dir)
		infof("Changing directory to %s", dir)
		cwd, err := os.Getwd()
		assert.NoError(t, err)
		ic.workDir = dir
		err = os.Chdir(dir)
		assert.NoError(t, err)
		defer os.Chdir(cwd)
		a(t, ic)
	}
}

// debugShell opens a new Terminal window in the test working directory.
func debugShell() action {
	return func(t testing.TB, ic testContext) {
		infof("Starting debug shell")
		err := ftlexec.Command(ic, log.Debug, ic.workDir, "open", "-n", "-W", "-a", "Terminal", ".").RunBuffered(ic)
		assert.NoError(t, err)
	}
}

// exec runs a command from the test working directory.
func exec(cmd string, args ...string) action {
	return func(t testing.TB, ic testContext) {
		infof("Executing: %s %s", cmd, shellquote.Join(args...))
		err := ftlexec.Command(ic, log.Debug, ic.workDir, cmd, args...).RunBuffered(ic)
		assert.NoError(t, err)
	}
}

// execWithOutput runs a command from the test working directory.
// The output is captured and is returned as part of the error.
func execWithOutput(cmd string, args ...string) action {
	return func(t testing.TB, ic testContext) {
		infof("Executing: %s %s", cmd, shellquote.Join(args...))
		output, err := ftlexec.Capture(ic, ic.workDir, cmd, args...)
		assert.NoError(t, err, "%s", string(output))
	}
}

// expectError wraps an action and expects it to return an error with the given message.
func expectError(action action, expectedErrorMsg string) action {
	return func(t testing.TB, ic testContext) {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(TestingError); ok {
					assert.Contains(t, string(e), expectedErrorMsg)
				} else {
					panic(r)
				}
			}
		}()
		action(t, ic)
	}
}

// Deploy a module from the working directory and wait for it to become available.
func deploy(module string) action {
	return chain(
		exec("ftl", "deploy", module),
		wait(module),
	)
}

// Build modules from the working directory and wait for it to become available.
func build(modules ...string) action {
	args := []string{"build"}
	args = append(args, modules...)
	return exec("ftl", args...)
}

// wait for the given module to deploy.
func wait(module string) action {
	return func(t testing.TB, ic testContext) {
		infof("Waiting for %s to become ready", module)
		// There's a bit of a bug here: wait() is already being retried by the
		// test harness, so in the error case we'll be waiting N^2 times. This
		// is fine for now, but we should fix this in the future.
		ic.AssertWithRetry(t, func(t testing.TB, ic testContext) {
			status, err := ic.controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
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

func sleep(duration time.Duration) action {
	return func(t testing.TB, ic testContext) {
		infof("Sleeping for %s", duration)
		time.Sleep(duration)
	}
}

// Assert that a file exists in the working directory.
func fileExists(path string) action {
	return func(t testing.TB, ic testContext) {
		infof("Checking that %s exists", path)
		_, err := os.Stat(filepath.Join(ic.workDir, path))
		assert.NoError(t, err)
	}
}

// Assert that a file exists and its content contains the given text.
//
// If "path" is relative it will be to the working directory.
func fileContains(path, needle string) action {
	return func(t testing.TB, ic testContext) {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		infof("Checking that the content of %s is correct", absPath)
		data, err := os.ReadFile(absPath)
		assert.NoError(t, err)
		actual := string(data)
		assert.Contains(t, actual, needle)
	}
}

// Assert that a file exists and its content is equal to the given text.
//
// If "path" is relative it will be to the working directory.
func fileContent(path, expected string) action {
	return func(t testing.TB, ic testContext) {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		infof("Checking that the content of %s is correct", absPath)
		data, err := os.ReadFile(absPath)
		assert.NoError(t, err)
		expected = strings.TrimSpace(expected)
		actual := strings.TrimSpace(string(data))
		assert.Equal(t, expected, actual)
	}
}

type obj map[string]any

// Call a verb.
//
// "check" may be nil
func call(module, verb string, request obj, check func(t testing.TB, response obj)) action {
	return func(t testing.TB, ic testContext) {
		infof("Calling %s.%s", module, verb)
		data, err := json.Marshal(request)
		assert.NoError(t, err)
		resp, err := ic.verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
			Verb: &schemapb.Ref{Module: module, Name: verb},
			Body: data,
		}))
		assert.NoError(t, err)
		var response obj
		assert.Zero(t, resp.Msg.GetError(), "verb failed: %s", resp.Msg.GetError().GetMessage())
		err = json.Unmarshal(resp.Msg.GetBody(), &response)
		assert.NoError(t, err)
		if check != nil {
			check(t, response)
		}
	}
}

// Fail expects the next action to fail.
func fail(next action, msg string, args ...any) action {
	return func(t testing.TB, ic testContext) {
		infof("Expecting failure of nested action")
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

// Query a single row from a database.
func queryRow(database string, query string, expected ...interface{}) action {
	return func(t testing.TB, ic testContext) {
		infof("Querying %s: %s", database, query)
		db, err := sql.Open("pgx", fmt.Sprintf("postgres://postgres:secret@localhost:54320/%s?sslmode=disable", database))
		assert.NoError(t, err)
		defer db.Close()
		actual := make([]any, len(expected))
		for i := range actual {
			actual[i] = new(any)
		}
		err = db.QueryRowContext(ic, query).Scan(actual...)
		assert.NoError(t, err)
		for i := range actual {
			actual[i] = *actual[i].(*any)
		}
		for i, a := range actual {
			assert.Equal(t, a, expected[i])
		}
	}
}

// Create a database for use by a module.
func createDBAction(module, dbName string, isTest bool) action {
	return func(t testing.TB, ic testContext) {
		createDB(t, module, dbName, isTest)
	}
}

func createDB(t testing.TB, module, dbName string, isTestDb bool) {
	// insert test suffix if needed when actually setting up db
	if isTestDb {
		dbName += "_test"
	}
	infof("Creating database %s", dbName)
	db, err := sql.Open("pgx", "postgres://postgres:secret@localhost:54320/ftl?sslmode=disable")
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
		// Terminate any dangling connections.
		_, err := db.Exec(`
				SELECT pid, pg_terminate_backend(pid)
				FROM pg_stat_activity
				WHERE datname = $1 AND pid <> pg_backend_pid()`,
			dbName)
		assert.NoError(t, err)
		_, err = db.Exec("DROP DATABASE " + dbName)
		assert.NoError(t, err)
	})
}

func connectToFTLDatabase(ctx context.Context, actionFunc func(conn *pgxpool.Conn) error) error {
	dsn := "postgres://postgres:secret@localhost:54320/ftl?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire PG PubSub connection: %w", err)
	}
	defer conn.Release()

	return actionFunc(conn)
}

type asyncCallRow struct {
	Verb              schema.RefKey
	Origin            string
	State             string
	ScheduledAt       time.Time
	Request           []byte
	Response          []byte
	Error             optional.Option[string]
	RemainingAttempts int32
	Backoff           time.Duration
	MaxBackoff        time.Duration
}

func getAsyncCalls(callback func([]asyncCallRow) error) action {
	return func(t testing.TB, ic testContext) error {
		return connectToFTLDatabase(ic.Context, func(conn *pgxpool.Conn) error {
			rows, err := conn.Query(ic.Context, `
		SELECT verb, origin, state, scheduled_at, request, response, error, remaining_attempts, backoff::interval, max_backoff::interval
		FROM async_calls
		ORDER BY created_at ASC`)
			if err != nil {
				return err
			}
			defer rows.Close()
			var items []asyncCallRow
			for rows.Next() {
				var i asyncCallRow
				if err := rows.Scan(
					&i.Verb,
					&i.Origin,
					&i.State,
					&i.ScheduledAt,
					&i.Request,
					&i.Response,
					&i.Error,
					&i.RemainingAttempts,
					&i.Backoff,
					&i.MaxBackoff,
				); err != nil {
					return err
				}
				items = append(items, i)
			}
			if err := rows.Err(); err != nil {
				return err
			}
			return callback(items)
		})
	}
}

func getFSMInstances(callback func([]gensql.FsmInstance) error) action {
	return func(t testing.TB, ic testContext) error {
		return connectToFTLDatabase(ic.Context, func(conn *pgxpool.Conn) error {
			rows, err := conn.Query(ic.Context, `
			SELECT id, created_at, fsm, key, status, current_state, destination_state, async_call_id
			FROM fsm_instances`)
			if err != nil {
				return err
			}
			defer rows.Close()
			var items []gensql.FsmInstance
			for rows.Next() {
				var i gensql.FsmInstance
				if err := rows.Scan(
					&i.ID,
					&i.CreatedAt,
					&i.Fsm,
					&i.Key,
					&i.Status,
					&i.CurrentState,
					&i.DestinationState,
					&i.AsyncCallID,
				); err != nil {
					return err
				}
				items = append(items, i)
			}
			if err := rows.Err(); err != nil {
				return err
			}
			return callback(items)
		})
	}
}

// Create a directory in the working directory
func mkdir(dir string) action {
	return func(t testing.TB, ic testContext) {
		infof("Creating directory %s", dir)
		err := os.MkdirAll(filepath.Join(ic.workDir, dir), 0700)
		assert.NoError(t, err)
	}
}

type httpResponse struct {
	status    int
	headers   map[string][]string
	jsonBody  map[string]any
	bodyBytes []byte
}

func jsonData(t testing.TB, body interface{}) []byte {
	b, err := json.Marshal(body)
	assert.NoError(t, err)
	return b
}

// httpCall makes an HTTP call to the running FTL ingress endpoint.
func httpCall(method string, path string, body []byte, onResponse func(t testing.TB, resp *httpResponse)) action {
	return func(t testing.TB, ic testContext) {
		infof("HTTP %s %s", method, path)
		baseURL, err := url.Parse(fmt.Sprintf("http://localhost:8892/ingress"))
		assert.NoError(t, err)

		r, err := http.NewRequestWithContext(ic, method, baseURL.JoinPath(path).String(), bytes.NewReader(body))
		assert.NoError(t, err)

		r.Header.Add("Content-Type", "application/json")

		client := http.Client{}
		resp, err := client.Do(r)
		assert.NoError(t, err)
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var resBody map[string]any
		// ignore the error here since some responses are just `[]byte`.
		_ = json.Unmarshal(bodyBytes, &resBody)

		onResponse(t, &httpResponse{
			status:    resp.StatusCode,
			headers:   resp.Header,
			jsonBody:  resBody,
			bodyBytes: bodyBytes,
		})
	}
}

// Run "go test" in the given module.
func testModule(module string) action {
	return chdir(module, exec("go", "test", "-v", "."))
}
