//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/integration"
)

func TestCron(t *testing.T) {
	dir := t.TempDir()
	// Due to some MacOS magic, /tmp differs between this test code and the
	// executing module, so we need to pass the file path as an environment
	// variable.
	tmpFile := filepath.Join(dir, "cron.txt")
	t.Setenv("DEST_FILE", tmpFile)

	t.Cleanup(func() { _ = os.Remove(tmpFile) })

	integration.Run(t, "",
		integration.CopyModule("cron"),
		integration.Deploy("cron"),
		func(t testing.TB, ic integration.TestContext) {
			_, err := os.Stat(tmpFile)
			assert.NoError(t, err)
		},
	)
}

func TestLifecycle(t *testing.T) {
	integration.Run(t, "",
		integration.Exec("ftl", "init", "go", ".", "echo"),
		integration.Deploy("echo"),
		integration.Call("echo", "echo", integration.Obj{"name": "Bob"}, func(t testing.TB, response integration.Obj) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
	)
}

func TestInterModuleCall(t *testing.T) {
	integration.Run(t, "",
		integration.CopyModule("echo"),
		integration.CopyModule("time"),
		integration.Deploy("time"),
		integration.Deploy("echo"),
		integration.Call("echo", "echo", integration.Obj{"name": "Bob"}, func(t testing.TB, response integration.Obj) {
			message, ok := response["message"].(string)
			assert.True(t, ok, "message is not a string: %s", repr.String(response))
			if !strings.HasPrefix(message, "Hello, Bob!!! It is ") {
				t.Fatalf("unexpected response: %q", response)
			}
		}),
	)
}

func TestNonExportedDecls(t *testing.T) {
	integration.Run(t, "",
		integration.CopyModule("time"),
		integration.Deploy("time"),
		integration.CopyModule("echo"),
		integration.Deploy("echo"),
		integration.CopyModule("notexportedverb"),
		integration.ExpectError(
			integration.ExecWithOutput("ftl", "deploy", "notexportedverb"),
			"call first argument must be a function but is an unresolved reference to echo.Echo, does it need to be exported?"),
	)
}

func TestUndefinedExportedDecls(t *testing.T) {
	integration.Run(t, "",
		integration.CopyModule("time"),
		integration.Deploy("time"),
		integration.CopyModule("echo"),
		integration.Deploy("echo"),
		integration.CopyModule("undefinedverb"),
		integration.ExpectError(
			integration.ExecWithOutput("ftl", "deploy", "undefinedverb"),
			"call first argument must be a function but is an unresolved reference to echo.Undefined"),
	)
}

func TestDatabase(t *testing.T) {
	integration.Run(t, "database/ftl-project.toml",
		// deploy real module against "testdb"
		integration.CopyModule("database"),
		integration.CreateDBAction("database", "testdb", false),
		integration.Deploy("database"),
		integration.Call("database", "insert", integration.Obj{"data": "hello"}, nil),
		integration.QueryRow("testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		integration.CreateDBAction("database", "testdb", true),
		integration.ExecModuleTest("database"),
		integration.QueryRow("testdb", "SELECT data FROM requests", "hello"),
	)
}

func TestSchemaGenerate(t *testing.T) {
	integration.Run(t, "",
		integration.CopyDir("../schema-generate", "schema-generate"),
		integration.Mkdir("build/schema-generate"),
		integration.Exec("ftl", "schema", "generate", "schema-generate", "build/schema-generate"),
		integration.FileContains("build/schema-generate/test.txt", "olleh"),
	)
}

func TestHttpEncodeOmitempty(t *testing.T) {
	integration.Run(t, "",
		integration.CopyModule("omitempty"),
		integration.Deploy("omitempty"),
		integration.HttpCall(http.MethodGet, "/get", integration.JsonData(t, integration.Obj{}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			_, ok := resp.JsonBody["mustset"]
			assert.True(t, ok)
			_, ok = resp.JsonBody["error"]
			assert.False(t, ok)
		}),
	)
}

func TestRuntimeReflection(t *testing.T) {
	integration.Run(t, "",
		integration.CopyModule("runtimereflection"),
		integration.ExecModuleTest("runtimereflection"),
	)
}

func TestModuleUnitTests(t *testing.T) {
	integration.Run(t, "",
		integration.CopyModule("time"),
		integration.CopyModule("wrapped"),
		integration.CopyModule("verbtypes"),
		integration.Build("time", "wrapped", "verbtypes"),
		integration.ExecModuleTest("wrapped"),
		integration.ExecModuleTest("verbtypes"),
	)
}

func TestLease(t *testing.T) {
	integration.Run(t, "",
		integration.CopyModule("leases"),
		integration.Build("leases"),
		// checks if leases work in a unit test environment
		integration.ExecModuleTest("leases"),
		integration.Deploy("leases"),
		// checks if it leases work with a real controller
		func(t testing.TB, ic integration.TestContext) {
			// Start a lease.
			wg := errgroup.Group{}
			wg.Go(func() error {
				integration.Infof("Acquiring lease")
				resp, err := ic.Verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
					Verb: &schemapb.Ref{Module: "leases", Name: "acquire"},
					Body: []byte("{}"),
				}))
				if respErr := resp.Msg.GetError(); respErr != nil {
					return fmt.Errorf("received error on first call: %v", respErr)
				}
				return err
			})

			time.Sleep(time.Second)

			integration.Infof("Trying to acquire lease again")
			// Trying to obtain the lease again should fail.
			resp, err := ic.Verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
				Verb: &schemapb.Ref{Module: "leases", Name: "acquire"},
				Body: []byte("{}"),
			}))
			assert.NoError(t, err)
			if resp.Msg.GetError() == nil || !strings.Contains(resp.Msg.GetError().Message, "could not acquire lease") {
				t.Fatalf("expected error but got: %#v", resp.Msg.GetError())
			}
			err = wg.Wait()
			assert.NoError(t, err)
		},
	)
}

func TestFSMGoTests(t *testing.T) {
	logFilePath := filepath.Join(t.TempDir(), "fsm.log")
	t.Setenv("FSM_LOG_FILE", logFilePath)
	integration.Run(t, "",
		integration.CopyModule("fsm"),
		integration.Build("fsm"),
		integration.ExecModuleTest("fsm"),
	)
}

func TestFSM(t *testing.T) {
	logFilePath := filepath.Join(t.TempDir(), "fsm.log")
	t.Setenv("FSM_LOG_FILE", logFilePath)
	fsmInState := func(instance, status, state string) integration.Action {
		return integration.QueryRow("ftl", fmt.Sprintf(`
			SELECT status, current_state
			FROM fsm_instances
			WHERE fsm = 'fsm.fsm' AND key = '%s'
		`, instance), status, state)
	}
	integration.Run(t, "",
		integration.CopyModule("fsm"),
		integration.Deploy("fsm"),

		integration.Call("fsm", "sendOne", integration.Obj{"instance": "1"}, nil),
		integration.Call("fsm", "sendOne", integration.Obj{"instance": "2"}, nil),
		integration.FileContains(logFilePath, "start 1"),
		integration.FileContains(logFilePath, "start 2"),
		fsmInState("1", "running", "fsm.start"),
		fsmInState("2", "running", "fsm.start"),

		integration.Call("fsm", "sendOne", integration.Obj{"instance": "1"}, nil),
		integration.FileContains(logFilePath, "middle 1"),
		fsmInState("1", "running", "fsm.middle"),

		integration.Call("fsm", "sendOne", integration.Obj{"instance": "1"}, nil),
		integration.FileContains(logFilePath, "end 1"),
		fsmInState("1", "completed", "fsm.end"),

		integration.Fail(integration.Call("fsm", "sendOne", integration.Obj{"instance": "1"}, nil),
			"FSM instance 1 is already in state fsm.end"),

		// Invalid state transition
		integration.Fail(integration.Call("fsm", "sendTwo", integration.Obj{"instance": "2"}, nil),
			"invalid state transition"),

		integration.Call("fsm", "sendOne", integration.Obj{"instance": "2"}, nil),
		integration.FileContains(logFilePath, "middle 2"),
		fsmInState("2", "running", "fsm.middle"),

		// Invalid state transition
		integration.Fail(integration.Call("fsm", "sendTwo", integration.Obj{"instance": "2"}, nil),
			"invalid state transition"),
	)
}

func TestFSMRetry(t *testing.T) {
	checkRetries := func(origin, verb string, delays []time.Duration) integration.Action {
		return func(t testing.TB, ic integration.TestContext) {
			results := []any{}
			for i := 0; i < len(delays); i++ {
				values := integration.GetRow(t, ic, "ftl", fmt.Sprintf("SELECT scheduled_at FROM async_calls WHERE origin = '%s' AND verb = '%s' AND state = 'error' ORDER BY created_at LIMIT 1 OFFSET %d", origin, verb, i), 1)
				results = append(results, values[0])
			}
			times := []time.Time{}
			for i, r := range results {
				ts, ok := r.(time.Time)
				assert.True(t, ok, "unexpected time value: %v", r)
				times = append(times, ts)
				if i > 0 {
					delay := times[i].Sub(times[i-1])
					targetDelay := delays[i-1]
					assert.True(t, delay >= targetDelay && delay < time.Second+targetDelay, "unexpected time diff for %s retry %d: %v (expected %v - %v)", origin, i, delay, targetDelay, time.Second+targetDelay)
				}
			}
		}
	}

	integration.Run(t, "",
		integration.CopyModule("fsmretry"),
		integration.Build("fsmretry"),
		integration.Deploy("fsmretry"),
		// start 2 FSM instances
		integration.Call("fsmretry", "start", integration.Obj{"id": "1"}, func(t testing.TB, response integration.Obj) {}),
		integration.Call("fsmretry", "start", integration.Obj{"id": "2"}, func(t testing.TB, response integration.Obj) {}),

		integration.Sleep(2*time.Second),

		// transition the FSM, should fail each time.
		integration.Call("fsmretry", "startTransitionToTwo", integration.Obj{"id": "1"}, func(t testing.TB, response integration.Obj) {}),
		integration.Call("fsmretry", "startTransitionToThree", integration.Obj{"id": "2"}, func(t testing.TB, response integration.Obj) {}),

		integration.Sleep(8*time.Second), //6s is longest run of retries

		// both FSMs instances should have failed
		integration.QueryRow("ftl", "SELECT COUNT(*) FROM fsm_instances WHERE status = 'failed'", int64(2)),

		integration.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s'", "fsm:fsmretry.fsm:1", "fsmretry.state2"), int64(4)),
		checkRetries("fsm:fsmretry.fsm:1", "fsmretry.state2", []time.Duration{time.Second, time.Second, time.Second}),
		integration.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s'", "fsm:fsmretry.fsm:2", "fsmretry.state3"), int64(4)),
		checkRetries("fsm:fsmretry.fsm:2", "fsmretry.state3", []time.Duration{time.Second, 2 * time.Second, 3 * time.Second}),
	)
}
