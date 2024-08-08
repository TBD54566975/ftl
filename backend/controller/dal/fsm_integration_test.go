//go:build integration

package dal_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	in "github.com/TBD54566975/ftl/integration"
	"github.com/alecthomas/assert/v2"
)

func TestFSM(t *testing.T) {
	logFilePath := filepath.Join(t.TempDir(), "fsm.log")
	t.Setenv("FSM_LOG_FILE", logFilePath)
	fsmInState := func(instance, status, state string) in.Action {
		return in.QueryRow("ftl", fmt.Sprintf(`
			SELECT status, current_state
			FROM fsm_instances
			WHERE fsm = 'fsm.fsm' AND key = '%s'
		`, instance), status, state)
	}
	in.Run(t, "",
		in.CopyModule("fsm"),
		in.Deploy("fsm"),

		in.Call[in.Obj, in.Obj]("fsm", "sendOne", in.Obj{"instance": "1"}, nil),
		in.Call[in.Obj, in.Obj]("fsm", "sendOne", in.Obj{"instance": "2"}, nil),
		in.FileContains(logFilePath, "start 1"),
		in.FileContains(logFilePath, "start 2"),
		fsmInState("1", "running", "fsm.start"),
		fsmInState("2", "running", "fsm.start"),

		in.Call[in.Obj, in.Obj]("fsm", "sendOne", in.Obj{"instance": "1"}, nil),
		in.FileContains(logFilePath, "middle 1"),
		fsmInState("1", "running", "fsm.middle"),

		in.Call[in.Obj, in.Obj]("fsm", "sendOne", in.Obj{"instance": "1"}, nil),
		in.FileContains(logFilePath, "end 1"),
		fsmInState("1", "completed", "fsm.end"),

		in.Fail(in.Call[in.Obj, in.Obj]("fsm", "sendOne", in.Obj{"instance": "1"}, nil),
			"FSM instance 1 is already in state fsm.end"),

		// Invalid state transition
		in.Fail(in.Call[in.Obj, in.Obj]("fsm", "sendTwo", in.Obj{"instance": "2"}, nil),
			"invalid state transition"),

		in.Call[in.Obj, in.Obj]("fsm", "sendOne", in.Obj{"instance": "2"}, nil),
		in.FileContains(logFilePath, "middle 2"),
		fsmInState("2", "running", "fsm.middle"),

		// Invalid state transition
		in.Fail(in.Call[in.Obj, in.Obj]("fsm", "sendTwo", in.Obj{"instance": "2"}, nil),
			"invalid state transition"),
	)
}

func TestFSMRetry(t *testing.T) {
	checkRetries := func(origin, verb string, delays []time.Duration) in.Action {
		return func(t testing.TB, ic in.TestContext) {
			results := []any{}
			for i := 0; i < len(delays); i++ {
				values := in.GetRow(t, ic, "ftl", fmt.Sprintf("SELECT scheduled_at FROM async_calls WHERE origin = '%s' AND verb = '%s' AND state = 'error' AND catching = false ORDER BY created_at LIMIT 1 OFFSET %d", origin, verb, i), 1)
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
					acceptableWindow := 1500 * time.Millisecond
					assert.True(t, delay >= targetDelay && delay < acceptableWindow+targetDelay, "unexpected time diff for %s retry %d: %v (expected %v - %v)", origin, i, delay, targetDelay, acceptableWindow+targetDelay)
				}
			}
		}
	}

	in.Run(t, "",
		in.CopyModule("fsmretry"),
		in.Build("fsmretry"),
		in.Deploy("fsmretry"),

		// start 3 FSM instances
		in.Call("fsmretry", "start", in.Obj{"id": "1"}, func(t testing.TB, response any) {}),
		in.Call("fsmretry", "start", in.Obj{"id": "2"}, func(t testing.TB, response any) {}),
		in.Call("fsmretry", "start", in.Obj{"id": "3"}, func(t testing.TB, response any) {}),

		in.Sleep(2*time.Second),

		// transition the FSM, should fail each time.
		in.Call("fsmretry", "startTransitionToTwo", in.Obj{"id": "1", "failCatch": false}, func(t testing.TB, response any) {}),
		in.Call("fsmretry", "startTransitionToThree", in.Obj{"id": "2"}, func(t testing.TB, response any) {}),
		in.Call("fsmretry", "startTransitionToTwo", in.Obj{"id": "3", "failCatch": true}, func(t testing.TB, response any) {}),

		in.Sleep(9*time.Second), //6s is longest run of retries

		// First two FSMs instances should have failed
		// Third one will not as it is still catching
		in.QueryRow("ftl", "SELECT COUNT(*) FROM fsm_instances WHERE status = 'failed'", int64(2)),

		// first FSM instance should have tried 3 times, and be caught once
		in.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s' AND catching = false", "fsm:fsmretry.fsm:1", "fsmretry.state2"), int64(3)),
		in.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s' AND catching = true", "fsm:fsmretry.fsm:1", "fsmretry.state2"), int64(1)),
		checkRetries("fsm:fsmretry.fsm:1", "fsmretry.state2", []time.Duration{2 * time.Second, 2 * time.Second}),

		// second FSM instance should have tried 3 times, and not be caught
		in.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s' AND catching = false", "fsm:fsmretry.fsm:2", "fsmretry.state3"), int64(3)),
		in.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s' AND catching = true", "fsm:fsmretry.fsm:2", "fsmretry.state3"), int64(0)),
		checkRetries("fsm:fsmretry.fsm:2", "fsmretry.state3", []time.Duration{2 * time.Second, 3 * time.Second}),

		// third FSM instance should have tried 3 times, and be caught indefinitely
		in.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s' AND catching = false", "fsm:fsmretry.fsm:3", "fsmretry.state2"), int64(3)),
		func(t testing.TB, ic in.TestContext) {
			counts := in.GetRow(t, ic, "ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s' AND catching = true", "fsm:fsmretry.fsm:3", "fsmretry.state2"), 1)
			assert.True(t, counts[0].(int64) >= 2, "expected at least 2 retries, got %d", counts[0].(int64))
		},
		checkRetries("fsm:fsmretry.fsm:1", "fsmretry.state2", []time.Duration{2 * time.Second, 2 * time.Second}),
	)
}

func TestFSMGoTests(t *testing.T) {
	logFilePath := filepath.Join(t.TempDir(), "fsm.log")
	t.Setenv("FSM_LOG_FILE", logFilePath)
	in.Run(t, "",
		in.CopyModule("fsm"),
		in.Build("fsm"),
		in.ExecModuleTest("fsm"),
	)
}
