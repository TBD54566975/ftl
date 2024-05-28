//go:build integration

package simple_test

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
)

func TestCron(t *testing.T) {
	dir := t.TempDir()
	// Due to some MacOS magic, /tmp differs between this test code and the
	// executing module, so we need to pass the file path as an environment
	// variable.
	tmpFile := filepath.Join(dir, "cron.txt")
	t.Setenv("DEST_FILE", tmpFile)

	t.Cleanup(func() { _ = os.Remove(tmpFile) })

	run(t, "",
		copyModule("cron"),
		deploy("cron"),
		func(t testing.TB, ic testContext) {
			_, err := os.Stat(tmpFile)
			assert.NoError(t, err)
		},
	)
}

func TestLifecycle(t *testing.T) {
	run(t, "",
		exec("ftl", "init", "go", ".", "echo"),
		deploy("echo"),
		call("echo", "echo", obj{"name": "Bob"}, func(t testing.TB, response obj) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
	)
}

func TestInterModuleCall(t *testing.T) {
	run(t, "",
		copyModule("echo"),
		copyModule("time"),
		deploy("time"),
		deploy("echo"),
		call("echo", "echo", obj{"name": "Bob"}, func(t testing.TB, response obj) {
			message, ok := response["message"].(string)
			assert.True(t, ok, "message is not a string: %s", repr.String(response))
			if !strings.HasPrefix(message, "Hello, Bob!!! It is ") {
				t.Fatalf("unexpected response: %q", response)
			}
		}),
	)
}

func TestNonExportedDecls(t *testing.T) {
	run(t, "",
		copyModule("time"),
		deploy("time"),
		copyModule("echo"),
		deploy("echo"),
		copyModule("notexportedverb"),
		expectError(execWithOutput("ftl", "deploy", "notexportedverb"), "call first argument must be a function but is an unresolved reference to echo.Echo, does it need to be exported?"),
	)
}

func TestUndefinedExportedDecls(t *testing.T) {
	run(t, "",
		copyModule("time"),
		deploy("time"),
		copyModule("echo"),
		deploy("echo"),
		copyModule("undefinedverb"),
		expectError(execWithOutput("ftl", "deploy", "undefinedverb"), "call first argument must be a function but is an unresolved reference to echo.Undefined"),
	)
}

func TestDatabase(t *testing.T) {
	run(t, "database/ftl-project.toml",
		// deploy real module against "testdb"
		copyModule("database"),
		createDBAction("database", "testdb", false),
		deploy("database"),
		call("database", "insert", obj{"data": "hello"}, nil),
		queryRow("testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		createDBAction("database", "testdb", true),
		testModule("database"),
		queryRow("testdb", "SELECT data FROM requests", "hello"),
	)
}

func TestSchemaGenerate(t *testing.T) {
	run(t, "",
		copyDir("../schema-generate", "schema-generate"),
		mkdir("build/schema-generate"),
		exec("ftl", "schema", "generate", "schema-generate", "build/schema-generate"),
		fileContains("build/schema-generate/test.txt", "olleh"),
	)
}

func TestHttpEncodeOmitempty(t *testing.T) {
	run(t, "",
		copyModule("omitempty"),
		deploy("omitempty"),
		httpCall(http.MethodGet, "/get", jsonData(t, obj{}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			_, ok := resp.jsonBody["mustset"]
			assert.True(t, ok)
			_, ok = resp.jsonBody["error"]
			assert.False(t, ok)
		}),
	)
}

func TestHttpIngress(t *testing.T) {
	run(t, "",
		copyModule("httpingress"),
		deploy("httpingress"),
		httpCall(http.MethodGet, "/users/123/posts/456", jsonData(t, obj{}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"Header from FTL"}, resp.headers["Get"])
			assert.Equal(t, []string{"application/json; charset=utf-8"}, resp.headers["Content-Type"])

			message, ok := resp.jsonBody["msg"].(string)
			assert.True(t, ok, "msg is not a string: %s", repr.String(resp.jsonBody))
			assert.Equal(t, "UserID: 123, PostID: 456", message)

			nested, ok := resp.jsonBody["nested"].(map[string]any)
			assert.True(t, ok, "nested is not a map: %s", repr.String(resp.jsonBody))
			goodStuff, ok := nested["good_stuff"].(string)
			assert.True(t, ok, "good_stuff is not a string: %s", repr.String(resp.jsonBody))
			assert.Equal(t, "This is good stuff", goodStuff)
		}),
		httpCall(http.MethodPost, "/users", jsonData(t, obj{"userId": 123, "postId": 345}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 201, resp.status)
			assert.Equal(t, []string{"Header from FTL"}, resp.headers["Post"])
			success, ok := resp.jsonBody["success"].(bool)
			assert.True(t, ok, "success is not a bool: %s", repr.String(resp.jsonBody))
			assert.True(t, success)
		}),
		// contains aliased field
		httpCall(http.MethodPost, "/users", jsonData(t, obj{"user_id": 123, "postId": 345}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 201, resp.status)
		}),
		httpCall(http.MethodPut, "/users/123", jsonData(t, obj{"postId": "346"}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"Header from FTL"}, resp.headers["Put"])
			assert.Equal(t, map[string]any{}, resp.jsonBody)
		}),
		httpCall(http.MethodDelete, "/users/123", jsonData(t, obj{}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"Header from FTL"}, resp.headers["Delete"])
			assert.Equal(t, map[string]any{}, resp.jsonBody)
		}),

		httpCall(http.MethodGet, "/html", jsonData(t, obj{}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"text/html; charset=utf-8"}, resp.headers["Content-Type"])
			assert.Equal(t, "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>", string(resp.bodyBytes))
		}),

		httpCall(http.MethodPost, "/bytes", []byte("Hello, World!"), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"application/octet-stream"}, resp.headers["Content-Type"])
			assert.Equal(t, []byte("Hello, World!"), resp.bodyBytes)
		}),

		httpCall(http.MethodGet, "/empty", nil, func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, nil, resp.headers["Content-Type"])
			assert.Equal(t, nil, resp.bodyBytes)
		}),

		httpCall(http.MethodGet, "/string", []byte("Hello, World!"), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.headers["Content-Type"])
			assert.Equal(t, []byte("Hello, World!"), resp.bodyBytes)
		}),

		httpCall(http.MethodGet, "/int", []byte("1234"), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.headers["Content-Type"])
			assert.Equal(t, []byte("1234"), resp.bodyBytes)
		}),
		httpCall(http.MethodGet, "/float", []byte("1234.56789"), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.headers["Content-Type"])
			assert.Equal(t, []byte("1234.56789"), resp.bodyBytes)
		}),
		httpCall(http.MethodGet, "/bool", []byte("true"), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.headers["Content-Type"])
			assert.Equal(t, []byte("true"), resp.bodyBytes)
		}),
		httpCall(http.MethodGet, "/error", nil, func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 500, resp.status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.headers["Content-Type"])
			assert.Equal(t, []byte("Error from FTL"), resp.bodyBytes)
		}),
		httpCall(http.MethodGet, "/array/string", jsonData(t, []string{"hello", "world"}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"application/json; charset=utf-8"}, resp.headers["Content-Type"])
			assert.Equal(t, jsonData(t, []string{"hello", "world"}), resp.bodyBytes)
		}),
		httpCall(http.MethodPost, "/array/data", jsonData(t, []obj{{"item": "a"}, {"item": "b"}}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"application/json; charset=utf-8"}, resp.headers["Content-Type"])
			assert.Equal(t, jsonData(t, []obj{{"item": "a"}, {"item": "b"}}), resp.bodyBytes)
		}),
		httpCall(http.MethodGet, "/typeenum", jsonData(t, obj{"name": "A", "value": "hello"}), func(t testing.TB, resp *httpResponse) {
			assert.Equal(t, 200, resp.status)
			assert.Equal(t, []string{"application/json; charset=utf-8"}, resp.headers["Content-Type"])
			assert.Equal(t, jsonData(t, obj{"name": "A", "value": "hello"}), resp.bodyBytes)
		}),
	)
}

func TestRuntimeReflection(t *testing.T) {
	run(t, "",
		copyModule("runtimereflection"),
		testModule("runtimereflection"),
	)
}

func TestModuleUnitTests(t *testing.T) {
	run(t, "",
		copyModule("time"),
		copyModule("wrapped"),
		copyModule("verbtypes"),
		build("time", "wrapped", "verbtypes"),
		testModule("wrapped"),
		testModule("verbtypes"),
	)
}

func TestLease(t *testing.T) {
	run(t, "",
		copyModule("leases"),
		build("leases"),
		// checks if leases work in a unit test environment
		testModule("leases"),
		deploy("leases"),
		// checks if it leases work with a real controller
		func(t testing.TB, ic testContext) {
			// Start a lease.
			wg := errgroup.Group{}
			wg.Go(func() error {
				infof("Acquiring lease")
				resp, err := ic.verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
					Verb: &schemapb.Ref{Module: "leases", Name: "acquire"},
					Body: []byte("{}"),
				}))
				if respErr := resp.Msg.GetError(); respErr != nil {
					return fmt.Errorf("received error on first call: %v", respErr)
				}
				return err
			})

			time.Sleep(time.Second)

			infof("Trying to acquire lease again")
			// Trying to obtain the lease again should fail.
			resp, err := ic.verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
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

func TestFSM(t *testing.T) {
	logFilePath := filepath.Join(t.TempDir(), "fsm.log")
	t.Setenv("FSM_LOG_FILE", logFilePath)
	fsmInState := func(instance, status, state string) action {
		return queryRow("ftl", fmt.Sprintf(`
			SELECT status, current_state
			FROM fsm_instances
			WHERE fsm = 'fsm.fsm' AND key = '%s'
		`, instance), status, state)
	}
	run(t, "",
		copyModule("fsm"),
		deploy("fsm"),

		call("fsm", "sendOne", obj{"instance": "1"}, nil),
		call("fsm", "sendOne", obj{"instance": "2"}, nil),
		fileContains(logFilePath, "start 1"),
		fileContains(logFilePath, "start 2"),
		fsmInState("1", "running", "fsm.start"),
		fsmInState("2", "running", "fsm.start"),

		call("fsm", "sendOne", obj{"instance": "1"}, nil),
		fileContains(logFilePath, "middle 1"),
		fsmInState("1", "running", "fsm.middle"),

		call("fsm", "sendOne", obj{"instance": "1"}, nil),
		fileContains(logFilePath, "end 1"),
		fsmInState("1", "completed", "fsm.end"),

		fail(call("fsm", "sendOne", obj{"instance": "1"}, nil),
			"FSM instance 1 is already in state fsm.end"),

		// Invalid state transition
		fail(call("fsm", "sendTwo", obj{"instance": "2"}, nil),
			"invalid state transition"),

		call("fsm", "sendOne", obj{"instance": "2"}, nil),
		fileContains(logFilePath, "middle 2"),
		fsmInState("2", "running", "fsm.middle"),

		// Invalid state transition
		fail(call("fsm", "sendTwo", obj{"instance": "2"}, nil),
			"invalid state transition"),
	)
}

func TestFSMRetry(t *testing.T) {
	retryCallQuery := func(origin string, verb string, offset int) string {
		return fmt.Sprintf("SELECT scheduled_at FROM async_calls WHERE origin = '%s' AND verb = '%s' AND state = 'error' ORDER BY created_at LIMIT 1 OFFSET %d", origin, verb, offset)
	}

	checkRetries := func(origin, verb string, delays []time.Duration) action {
		return func(t testing.TB, ic testContext) error {
			results := []any{}
			for i := 0; i < len(delays); i++ {
				values, err := getRow(ic, "ftl", retryCallQuery(origin, verb, i), 1)
				if err != nil {
					return err
				}
				results = append(results, values[0])
			}

			times := []time.Time{}
			for i, r := range results {
				t, ok := r.(time.Time)
				if !ok {
					return fmt.Errorf("unexpected time value: %v", r)
				}
				times = append(times, t)
				if i > 0 {
					delay := times[i].Sub(times[i-1])
					if delay < delays[i-1] || delay >= time.Second+delays[i-1] {
						return fmt.Errorf("unexpected time diff for %s retry %d: %v", origin, i, delay)
					}
				}
			}
			return nil
		}
	}

	run(t, "",
		copyModule("fsmretry"),
		build("fsmretry"),
		deploy("fsmretry"),
		// start 2 FSM instances
		call("fsmretry", "start", obj{"id": "1"}, func(response obj) error { return nil }),
		call("fsmretry", "start", obj{"id": "2"}, func(response obj) error { return nil }),

		sleep(2*time.Second),

		// transition the FSM, should fail each time.
		call("fsmretry", "startTransitionToTwo", obj{"id": "1"}, func(response obj) error { return nil }),
		call("fsmretry", "startTransitionToThree", obj{"id": "2"}, func(response obj) error { return nil }),

		sleep(8*time.Second), //6s is longest run of retries

		// both FSMs instances should have failed
		queryRow("ftl", "SELECT COUNT(*) FROM fsm_instances WHERE status = 'failed'", int64(2)),

		queryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s'", "fsm:fsmretry.fsm:1", "fsmretry.state2"), int64(4)),
		checkRetries("fsm:fsmretry.fsm:1", "fsmretry.state2", []time.Duration{time.Second, time.Second, time.Second}),
		queryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s'", "fsm:fsmretry.fsm:2", "fsmretry.state3"), int64(4)),
		checkRetries("fsm:fsmretry.fsm:2", "fsmretry.state3", []time.Duration{time.Second, 2 * time.Second, 3 * time.Second}),
	)
}
