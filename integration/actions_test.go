//go:build integration

package simple_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
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
	"github.com/alecthomas/assert/v2"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver
	"github.com/kballard/go-shellquote"
	"github.com/otiai10/copy"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	ftlexec "github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/scaffolder"
)

// scaffold a directory relative to the testdata directory to a directory relative to the working directory.
func scaffold(src, dest string, tmplCtx any) action {
	return func(t testing.TB, ic testContext) error {
		infof("Scaffolding %s -> %s", src, dest)
		return scaffolder.Scaffold(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest), tmplCtx)
	}
}

// Copy a module from the testdata directory to the working directory.
//
// Ensures that replace directives are correctly handled.
func copyModule(module string) action {
	return chain(
		copyDir(module, module),
		func(t testing.TB, ic testContext) error {
			return ftlexec.Command(ic, log.Debug, filepath.Join(ic.workDir, module), "go", "mod", "edit", "-replace", "github.com/TBD54566975/ftl="+ic.rootDir).RunBuffered(ic)
		},
	)
}

// Copy a directory from the testdata directory to the working directory.
func copyDir(src, dest string) action {
	return func(t testing.TB, ic testContext) error {
		infof("Copying %s -> %s", src, dest)
		return copy.Copy(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest))
	}
}

// chain multiple actions together.
func chain(actions ...action) action {
	return func(t testing.TB, ic testContext) error {
		for _, action := range actions {
			err := action(t, ic)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// chdir changes the test working directory to the subdirectory for the duration of the action.
func chdir(dir string, a action) action {
	return func(t testing.TB, ic testContext) error {
		dir := filepath.Join(ic.workDir, dir)
		infof("Changing directory to %s", dir)
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		ic.workDir = dir
		err = os.Chdir(dir)
		if err != nil {
			return err
		}
		defer os.Chdir(cwd)
		return a(t, ic)
	}
}

// debugShell opens a new Terminal window in the test working directory.
func debugShell() action {
	return func(t testing.TB, ic testContext) error {
		infof("Starting debug shell")
		return ftlexec.Command(ic, log.Debug, ic.workDir, "open", "-n", "-W", "-a", "Terminal", ".").RunBuffered(ic)
	}
}

// exec runs a command from the test working directory.
func exec(cmd string, args ...string) action {
	return func(t testing.TB, ic testContext) error {
		infof("Executing: %s %s", cmd, shellquote.Join(args...))
		err := ftlexec.Command(ic, log.Debug, ic.workDir, cmd, args...).RunBuffered(ic)
		if err != nil {
			return err
		}
		return nil
	}
}

// execWithOutput runs a command from the test working directory.
// The output is captured and is returned as part of the error.
func execWithOutput(cmd string, args ...string) action {
	return func(t testing.TB, ic testContext) error {
		infof("Executing: %s %s", cmd, shellquote.Join(args...))
		output, err := ftlexec.Capture(ic, ic.workDir, cmd, args...)
		if err != nil {
			return fmt.Errorf("command execution failed: %s, output: %s", err, string(output))
		}
		return nil
	}
}

// expectError wraps an action and expects it to return an error with the given message.
func expectError(action action, expectedErrorMsg string) action {
	return func(t testing.TB, ic testContext) error {
		err := action(t, ic)
		if err == nil {
			return fmt.Errorf("expected error %q, but got nil", expectedErrorMsg)
		}
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			return fmt.Errorf("expected error %q, but got %q", expectedErrorMsg, err.Error())
		}
		return nil
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
	return func(t testing.TB, ic testContext) error {
		infof("Waiting for %s to become ready", module)
		ic.AssertWithRetry(t, func(t testing.TB, ic testContext) error {
			status, err := ic.controller.Status(ic, connect.NewRequest(&ftlv1.StatusRequest{}))
			if err != nil {
				return err
			}
			for _, deployment := range status.Msg.Deployments {
				if deployment.Name == module {
					return nil
				}
			}
			return fmt.Errorf("deployment of module %q not found", module)
		})
		return nil
	}
}

func sleep(duration time.Duration) action {
	return func(t testing.TB, ic testContext) error {
		infof("Sleeping for %s", duration)
		time.Sleep(duration)
		return nil
	}
}

// Assert that a file exists in the working directory.
func fileExists(path string) action {
	return func(t testing.TB, ic testContext) error {
		infof("Checking that %s exists", path)
		_, err := os.Stat(filepath.Join(ic.workDir, path))
		return err
	}
}

// Assert that a file exists and its content contains the given text.
//
// If "path" is relative it will be to the working directory.
func fileContains(path, needle string) action {
	return func(t testing.TB, ic testContext) error {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		infof("Checking that the content of %s is correct", absPath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}
		actual := string(data)
		if !strings.Contains(actual, needle) {
			return fmt.Errorf("expected %q to contain %q", actual, needle)
		}
		return nil
	}
}

// Assert that a file exists and its content is equal to the given text.
//
// If "path" is relative it will be to the working directory.
func fileContent(path, expected string) action {
	return func(t testing.TB, ic testContext) error {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(ic.workDir, path)
		}
		infof("Checking that the content of %s is correct", absPath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}
		expected = strings.TrimSpace(expected)
		actual := strings.TrimSpace(string(data))
		if actual != expected {
			return errors.New(assert.Diff(expected, actual))
		}
		return nil
	}
}

type obj map[string]any

// Call a verb.
func call(module, verb string, request obj, check func(response obj) error) action {
	return func(t testing.TB, ic testContext) error {
		infof("Calling %s.%s", module, verb)
		data, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		resp, err := ic.verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
			Verb: &schemapb.Ref{Module: module, Name: verb},
			Body: data,
		}))
		if err != nil {
			return fmt.Errorf("failed to call verb: %w", err)
		}
		var response obj
		if resp.Msg.GetError() != nil {
			return fmt.Errorf("verb failed: %s", resp.Msg.GetError().GetMessage())
		}
		err = json.Unmarshal(resp.Msg.GetBody(), &response)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return check(response)
	}
}

// Fail expects the next action to fail.
func fail(next action, msg string, args ...any) action {
	return func(t testing.TB, ic testContext) error {
		err := next(t, ic)
		if err == nil {
			return fmt.Errorf("expected action to fail: "+msg, args...)
		}
		return nil
	}
}

// Query a single row from a database.
func queryRow(database string, query string, expected ...interface{}) action {
	return func(t testing.TB, ic testContext) error {
		infof("Querying %s: %s", database, query)
		db, err := sql.Open("pgx", fmt.Sprintf("postgres://postgres:secret@localhost:54320/%s?sslmode=disable", database))
		if err != nil {
			return err
		}
		defer db.Close()
		actual := make([]any, len(expected))
		for i := range actual {
			actual[i] = new(any)
		}
		err = db.QueryRowContext(ic, query).Scan(actual...)
		if err != nil {
			return err
		}
		for i := range actual {
			actual[i] = *actual[i].(*any)
		}
		for i, a := range actual {
			if a != expected[i] {
				return fmt.Errorf("%s:\n%s", query, assert.Diff(expected, actual))
			}
		}
		return nil
	}
}

// Create a database for use by a module.
func createDBAction(module, dbName string, isTest bool) action {
	return func(t testing.TB, ic testContext) error {
		createDB(t, module, dbName, isTest)
		return nil
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

// Create a directory in the working directory
func mkdir(dir string) action {
	return func(t testing.TB, ic testContext) error {
		infof("Creating directory %s", dir)
		return os.MkdirAll(filepath.Join(ic.workDir, dir), 0700)
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
func httpCall(method string, path string, body []byte, onResponse func(resp *httpResponse) error) action {
	return func(t testing.TB, ic testContext) error {
		infof("HTTP %s %s", method, path)
		baseURL, err := url.Parse(fmt.Sprintf("http://localhost:8892/ingress"))
		if err != nil {
			return err
		}

		r, err := http.NewRequestWithContext(ic, method, baseURL.JoinPath(path).String(), bytes.NewReader(body))
		if err != nil {
			return err
		}

		r.Header.Add("Content-Type", "application/json")

		client := http.Client{}
		resp, err := client.Do(r)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var resBody map[string]any
		// ignore the error here since some responses are just `[]byte`.
		_ = json.Unmarshal(bodyBytes, &resBody)

		return onResponse(&httpResponse{
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
