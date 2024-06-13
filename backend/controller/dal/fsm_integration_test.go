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

		in.Call("fsm", "sendOne", in.Obj{"instance": "1"}, nil),
		in.Call("fsm", "sendOne", in.Obj{"instance": "2"}, nil),
		in.FileContains(logFilePath, "start 1"),
		in.FileContains(logFilePath, "start 2"),
		fsmInState("1", "running", "fsm.start"),
		fsmInState("2", "running", "fsm.start"),

		in.Call("fsm", "sendOne", in.Obj{"instance": "1"}, nil),
		in.FileContains(logFilePath, "middle 1"),
		fsmInState("1", "running", "fsm.middle"),

		in.Call("fsm", "sendOne", in.Obj{"instance": "1"}, nil),
		in.FileContains(logFilePath, "end 1"),
		fsmInState("1", "completed", "fsm.end"),

		in.Fail(in.Call("fsm", "sendOne", in.Obj{"instance": "1"}, nil),
			"FSM instance 1 is already in state fsm.end"),

		// Invalid state transition
		in.Fail(in.Call("fsm", "sendTwo", in.Obj{"instance": "2"}, nil),
			"invalid state transition"),

		in.Call("fsm", "sendOne", in.Obj{"instance": "2"}, nil),
		in.FileContains(logFilePath, "middle 2"),
		fsmInState("2", "running", "fsm.middle"),

		// Invalid state transition
		in.Fail(in.Call("fsm", "sendTwo", in.Obj{"instance": "2"}, nil),
			"invalid state transition"),
	)
}

func TestFSMRetry(t *testing.T) {
	checkRetries := func(origin, verb string, delays []time.Duration) in.Action {
		return func(t testing.TB, ic in.TestContext) {
			results := []any{}
			for i := 0; i < len(delays); i++ {
				values := in.GetRow(t, ic, "ftl", fmt.Sprintf("SELECT scheduled_at FROM async_calls WHERE origin = '%s' AND verb = '%s' AND state = 'error' ORDER BY created_at LIMIT 1 OFFSET %d", origin, verb, i), 1)
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
		// start 2 FSM instances
		in.Call("fsmretry", "start", in.Obj{"id": "1"}, func(t testing.TB, response in.Obj) {}),
		in.Call("fsmretry", "start", in.Obj{"id": "2"}, func(t testing.TB, response in.Obj) {}),

		in.Sleep(2*time.Second),

		// transition the FSM, should fail each time.
		in.Call("fsmretry", "startTransitionToTwo", in.Obj{"id": "1"}, func(t testing.TB, response in.Obj) {}),
		in.Call("fsmretry", "startTransitionToThree", in.Obj{"id": "2"}, func(t testing.TB, response in.Obj) {}),

		in.Sleep(8*time.Second), //5s is longest run of retries

		// both FSMs instances should have failed
		in.QueryRow("ftl", "SELECT COUNT(*) FROM fsm_instances WHERE status = 'failed'", int64(2)),

		in.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s'", "fsm:fsmretry.fsm:1", "fsmretry.state2"), int64(3)),
		checkRetries("fsm:fsmretry.fsm:1", "fsmretry.state2", []time.Duration{2 * time.Second, 2 * time.Second}),
		in.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE origin = '%s' AND verb = '%s'", "fsm:fsmretry.fsm:2", "fsmretry.state3"), int64(3)),
		checkRetries("fsm:fsmretry.fsm:2", "fsmretry.state3", []time.Duration{2 * time.Second, 3 * time.Second}),
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
