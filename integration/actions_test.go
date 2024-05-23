//go:build integration

package simple_test

import (
	"bytes"
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

// Scaffold a directory relative to the testdata directory to a directory relative to the working directory.
func Scaffold(src, dest string, tmplCtx any) Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Scaffolding %s -> %s", src, dest)
		return scaffolder.Scaffold(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest), tmplCtx)
	}
}

// Copy a module from the testdata directory to the working directory.
//
// Ensures that replace directives are correctly handled.
func CopyModule(module string) Action {
	return Chain(
		CopyDir(module, module),
		func(t testing.TB, ic TestContext) error {
			return ftlexec.Command(ic, log.Debug, filepath.Join(ic.workDir, module), "go", "mod", "edit", "-replace", "github.com/TBD54566975/ftl="+ic.rootDir).RunBuffered(ic)
		},
	)
}

// Copy a directory from the testdata directory to the working directory.
func CopyDir(src, dest string) Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Copying %s -> %s", src, dest)
		return copy.Copy(filepath.Join(ic.testData, src), filepath.Join(ic.workDir, dest))
	}
}

// Chain multiple actions together.
func Chain(actions ...Action) Action {
	return func(t testing.TB, ic TestContext) error {
		for _, action := range actions {
			err := action(t, ic)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// Chdir changes the test working directory to the subdirectory for the duration of the action.
func Chdir(dir string, a Action) Action {
	return func(t testing.TB, ic TestContext) error {
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

// DebugShell opens a new Terminal window in the test working directory.
func DebugShell() Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Starting debug shell")
		return ftlexec.Command(ic, log.Debug, ic.workDir, "open", "-n", "-W", "-a", "Terminal", ".").RunBuffered(ic)
	}
}

// Exec runs a command from the test working directory.
func Exec(cmd string, args ...string) Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Executing: %s %s", cmd, shellquote.Join(args...))
		err := ftlexec.Command(ic, log.Debug, ic.workDir, cmd, args...).RunBuffered(ic)
		if err != nil {
			return err
		}
		return nil
	}
}

// ExecWithOutput runs a command from the test working directory.
// The output is captured and is returned as part of the error.
func ExecWithOutput(cmd string, args ...string) Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Executing: %s %s", cmd, shellquote.Join(args...))
		output, err := ftlexec.Capture(ic, ic.workDir, cmd, args...)
		if err != nil {
			return fmt.Errorf("command execution failed: %s, output: %s", err, string(output))
		}
		return nil
	}
}

// ExpectError wraps an action and expects it to return an error with the given message.
func ExpectError(action Action, expectedErrorMsg string) Action {
	return func(t testing.TB, ic TestContext) error {
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
func Deploy(module string) Action {
	return Chain(
		Exec("ftl", "deploy", module),
		Wait(module),
	)
}

// Build modules from the working directory and wait for it to become available.
func Build(modules ...string) Action {
	args := []string{"build"}
	args = append(args, modules...)
	return Exec("ftl", args...)
}

// Wait for the given module to deploy.
func Wait(module string) Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Waiting for %s to become ready", module)
		ic.AssertWithRetry(t, func(t testing.TB, ic TestContext) error {
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

func Sleep(duration time.Duration) Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Sleeping for %s", duration)
		time.Sleep(duration)
		return nil
	}
}

// Assert that a file exists in the working directory.
func FileExists(path string) Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Checking that %s exists", path)
		_, err := os.Stat(filepath.Join(ic.workDir, path))
		return err
	}
}

// Assert that a file exists in the working directory and contains the given text.
func FileContains(path, content string) Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Checking that %s contains %q", path, content)
		data, err := os.ReadFile(filepath.Join(ic.workDir, path))
		if err != nil {
			return err
		}
		if !strings.Contains(string(data), content) {
			return fmt.Errorf("%q not found in %q", content, string(data))
		}
		return nil
	}
}

type obj map[string]any

// Call a verb.
func Call(module, verb string, request obj, check func(response obj) error) Action {
	return func(t testing.TB, ic TestContext) error {
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

// Query a single row from a database.
func QueryRow(database string, query string, expected ...interface{}) Action {
	return func(t testing.TB, ic TestContext) error {
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
				return fmt.Errorf("expected %v, got %v", expected, actual)
			}
		}
		return nil
	}
}

// Create a database for use by a module.
func CreateDBAction(module, dbName string, isTest bool) Action {
	return func(t testing.TB, ic TestContext) error {
		CreateDB(t, module, dbName, isTest)
		return nil
	}
}

func CreateDB(t testing.TB, module, dbName string, isTestDb bool) {
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
func Mkdir(dir string) Action {
	return func(t testing.TB, ic TestContext) error {
		infof("Creating directory %s", dir)
		return os.MkdirAll(filepath.Join(ic.workDir, dir), 0700)
	}
}

type HttpResponse struct {
	status    int
	headers   map[string][]string
	jsonBody  map[string]any
	bodyBytes []byte
}

func JsonData(t testing.TB, body interface{}) []byte {
	b, err := json.Marshal(body)
	assert.NoError(t, err)
	return b
}

// HttpCall makes an HTTP call to the running FTL ingress endpoint.
func HttpCall(method string, path string, body []byte, onResponse func(resp *HttpResponse) error) Action {
	return func(t testing.TB, ic TestContext) error {
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

		return onResponse(&HttpResponse{
			status:    resp.StatusCode,
			headers:   resp.Header,
			jsonBody:  resBody,
			bodyBytes: bodyBytes,
		})
	}
}

func ExecModuleTest(module string) Action {
	return Chdir(module, Exec("go", "test", "-v", "."))
}
