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
	. "github.com/TBD54566975/ftl/integration"
)

func TestCron(t *testing.T) {
	dir := t.TempDir()
	// Due to some MacOS magic, /tmp differs between this test code and the
	// executing module, so we need to pass the file path as an environment
	// variable.
	tmpFile := filepath.Join(dir, "cron.txt")
	t.Setenv("DEST_FILE", tmpFile)

	t.Cleanup(func() { _ = os.Remove(tmpFile) })

	Run(t, "",
		CopyModule("cron"),
		Deploy("cron"),
		func(t testing.TB, ic TestContext) {
			_, err := os.Stat(tmpFile)
			assert.NoError(t, err)
		},
	)
}

func TestLifecycle(t *testing.T) {
	Run(t, "",
		Exec("ftl", "init", "go", ".", "echo"),
		Deploy("echo"),
		Call("echo", "echo", Obj{"name": "Bob"}, func(t testing.TB, response Obj) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
	)
}

func TestInterModuleCall(t *testing.T) {
	Run(t, "",
		CopyModule("echo"),
		CopyModule("time"),
		Deploy("time"),
		Deploy("echo"),
		Call("echo", "echo", Obj{"name": "Bob"}, func(t testing.TB, response Obj) {
			message, ok := response["message"].(string)
			assert.True(t, ok, "message is not a string: %s", repr.String(response))
			if !strings.HasPrefix(message, "Hello, Bob!!! It is ") {
				t.Fatalf("unexpected response: %q", response)
			}
		}),
	)
}

func TestNonExportedDecls(t *testing.T) {
	Run(t, "",
		CopyModule("time"),
		Deploy("time"),
		CopyModule("echo"),
		Deploy("echo"),
		CopyModule("notexportedverb"),
		ExpectError(
			ExecWithOutput("ftl", "deploy", "notexportedverb"),
			"call first argument must be a function but is an unresolved reference to echo.Echo, does it need to be exported?"),
	)
}

func TestUndefinedExportedDecls(t *testing.T) {
	Run(t, "",
		CopyModule("time"),
		Deploy("time"),
		CopyModule("echo"),
		Deploy("echo"),
		CopyModule("undefinedverb"),
		ExpectError(
			ExecWithOutput("ftl", "deploy", "undefinedverb"),
			"call first argument must be a function but is an unresolved reference to echo.Undefined"),
	)
}

func TestDatabase(t *testing.T) {
	Run(t, "database/ftl-project.toml",
		// deploy real module against "testdb"
		CopyModule("database"),
		CreateDBAction("database", "testdb", false),
		Deploy("database"),
		Call("database", "insert", Obj{"data": "hello"}, nil),
		QueryRow("testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		CreateDBAction("database", "testdb", true),
		ExecModuleTest("database"),
		QueryRow("testdb", "SELECT data FROM requests", "hello"),
	)
}

func TestSchemaGenerate(t *testing.T) {
	Run(t, "",
		CopyDir("../schema-generate", "schema-generate"),
		Mkdir("build/schema-generate"),
		Exec("ftl", "schema", "generate", "schema-generate", "build/schema-generate"),
		FileContains("build/schema-generate/test.txt", "olleh"),
	)
}

func TestHttpEncodeOmitempty(t *testing.T) {
	Run(t, "",
		CopyModule("omitempty"),
		Deploy("omitempty"),
		HttpCall(http.MethodGet, "/get", JsonData(t, Obj{}), func(t testing.TB, resp *HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			_, ok := resp.JsonBody["mustset"]
			assert.True(t, ok)
			_, ok = resp.JsonBody["error"]
			assert.False(t, ok)
		}),
	)
}

func TestRuntimeReflection(t *testing.T) {
	Run(t, "",
		CopyModule("runtimereflection"),
		ExecModuleTest("runtimereflection"),
	)
}

func TestModuleUnitTests(t *testing.T) {
	Run(t, "",
		CopyModule("time"),
		CopyModule("wrapped"),
		CopyModule("verbtypes"),
		Build("time", "wrapped", "verbtypes"),
		ExecModuleTest("wrapped"),
		ExecModuleTest("verbtypes"),
	)
}

func TestLease(t *testing.T) {
	Run(t, "",
		CopyModule("leases"),
		Build("leases"),
		// checks if leases work in a unit test environment
		ExecModuleTest("leases"),
		Deploy("leases"),
		// checks if it leases work with a real controller
		func(t testing.TB, ic TestContext) {
			// Start a lease.
			wg := errgroup.Group{}
			wg.Go(func() error {
				Infof("Acquiring lease")
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

			Infof("Trying to acquire lease again")
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
	Run(t, "",
		CopyModule("fsm"),
		Build("fsm"),
		ExecModuleTest("fsm"),
	)
}

func TestFSM(t *testing.T) {
	logFilePath := filepath.Join(t.TempDir(), "fsm.log")
	t.Setenv("FSM_LOG_FILE", logFilePath)
	fsmInState := func(instance, status, state string) Action {
		return QueryRow("ftl", fmt.Sprintf(`
			SELECT status, current_state
			FROM fsm_instances
			WHERE fsm = 'fsm.fsm' AND key = '%s'
		`, instance), status, state)
	}
	Run(t, "",
		CopyModule("fsm"),
		Deploy("fsm"),

		Call("fsm", "sendOne", Obj{"instance": "1"}, nil),
		Call("fsm", "sendOne", Obj{"instance": "2"}, nil),
		FileContains(logFilePath, "start 1"),
		FileContains(logFilePath, "start 2"),
		fsmInState("1", "running", "fsm.start"),
		fsmInState("2", "running", "fsm.start"),

		Call("fsm", "sendOne", Obj{"instance": "1"}, nil),
		FileContains(logFilePath, "middle 1"),
		fsmInState("1", "running", "fsm.middle"),

		Call("fsm", "sendOne", Obj{"instance": "1"}, nil),
		FileContains(logFilePath, "end 1"),
		fsmInState("1", "completed", "fsm.end"),

		Fail(Call("fsm", "sendOne", Obj{"instance": "1"}, nil),
			"FSM instance 1 is already in state fsm.end"),

		// Invalid state transition
		Fail(Call("fsm", "sendTwo", Obj{"instance": "2"}, nil),
			"invalid state transition"),

		Call("fsm", "sendOne", Obj{"instance": "2"}, nil),
		FileContains(logFilePath, "middle 2"),
		fsmInState("2", "running", "fsm.middle"),

		// Invalid state transition
		Fail(Call("fsm", "sendTwo", Obj{"instance": "2"}, nil),
			"invalid state transition"),
	)
}

func TestFSMRetry(t *testing.T) {
	checkRetries := func(origin, verb string, delays []time.Duration) Action {
		return func(t testing.TB, ic TestContext) {
			results := []any{}
			for i := 0; i < len(delays); i++ {
				values := GetRow(t, ic, "ftl", fmt.Sprintf("SELECT scheduled_at FROM async_calls WHERE origin = '%s' AND verb = '%s' AND state = 'error' ORDER BY created_at LIMIT 1 OFFSET %d", origin, verb, i), 1)
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

	Run(t, "",
		CopyModule("fsmretry"),
		Build("fsmretry"),
		Deploy("fsmretry"),
		// start 2 FSM instances
		Call("fsmretry", "start", Obj{"id": "1"}, func(t testing.TB, response Obj) {}),
		Call("fsmretry", "start", Obj{"id": "2"}, func(t testing.TB, response Obj) {}),

		Sleep(2*time.Second),

		// transition the FSM, should fail each time.
		Call("fsmretry", "startTransitionToTwo", Obj{"id": "1"}, func(t testing.TB, response Obj) {}),
		Call("fsmretry", "startTransitionToThree", Obj{"id": "2"}, func(t testing.TB, response Obj) {}),

		Sleep(8*time.Second), //6s is longest run of retries

		// both FSMs instances should have failed
		QueryRow("ftl", "SELECT COUNT(*) FROM fsm_instances WHERE status = 'failed'", int64(2)),

		QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s'", "fsm:fsmretry.fsm:1", "fsmretry.state2"), int64(4)),
		checkRetries("fsm:fsmretry.fsm:1", "fsmretry.state2", []time.Duration{time.Second, time.Second, time.Second}),
		QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s'", "fsm:fsmretry.fsm:2", "fsmretry.state3"), int64(4)),
		checkRetries("fsm:fsmretry.fsm:2", "fsmretry.state3", []time.Duration{time.Second, 2 * time.Second, 3 * time.Second}),
	)
}
