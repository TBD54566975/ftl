//go:build integration

package watch

import (
	"context" //nolint:depguard
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/pubsub"

	in "github.com/block/ftl/internal/integration"
	"github.com/block/ftl/internal/moduleconfig"
)

const pollFrequency = time.Millisecond * 500

func TestWatch(t *testing.T) {
	var events chan WatchEvent
	var topic *pubsub.Topic[WatchEvent]
	var one, two moduleconfig.UnvalidatedModuleConfig

	w := NewWatcher("**/*.go", "go.mod", "go.sum")
	in.Run(t,
		func(tb testing.TB, ic in.TestContext) {
			events, topic = startWatching(ic, t, w, ic.WorkingDir())
		},
		// Add modules
		in.FtlNew("go", "one"),
		in.FtlNew("go", "two"),
		func(tb testing.TB, ic in.TestContext) {
			one = loadModule(t, ic.WorkingDir(), "one")
			two = loadModule(t, ic.WorkingDir(), "two")
		},
		func(tb testing.TB, ic in.TestContext) {
			// Module creation may insert the files and then modify them (eg `go mod tidy`)
			// so we allow change events to occur
			waitForEventsWhileIgnoringEvents(tb, events, []WatchEvent{
				WatchEventModuleAdded{Config: one},
				WatchEventModuleAdded{Config: two},
			}, []WatchEvent{
				WatchEventModuleChanged{Config: one},
				WatchEventModuleChanged{Config: two},
			})
		},

		// Delete and modify a module
		in.RemoveDir("two"),
		updateModFile("one"),
		func(tb testing.TB, ic in.TestContext) {
			waitForEvents(tb, events, []WatchEvent{
				WatchEventModuleChanged{Config: one},
				WatchEventModuleRemoved{Config: two},
			})
		},

		// Cleanup
		func(tb testing.TB, ic in.TestContext) {
			topic.Close()
		},
	)
}

func TestWatchWithBuildModifyingFiles(t *testing.T) {
	var events chan WatchEvent
	var topic *pubsub.Topic[WatchEvent]
	var transaction ModifyFilesTransaction
	w := NewWatcher("**/*.go", "go.mod", "go.sum")

	in.Run(t,
		func(tb testing.TB, ic in.TestContext) {
			events, topic = startWatching(ic, t, w, ic.WorkingDir())
		},

		in.FtlNew("go", "one"),
		func(tb testing.TB, ic in.TestContext) {
			waitForEventsWhileIgnoringEvents(tb, events, []WatchEvent{
				WatchEventModuleAdded{Config: loadModule(t, ic.WorkingDir(), "one")},
			}, []WatchEvent{
				WatchEventModuleChanged{Config: loadModule(t, ic.WorkingDir(), "one")},
			})
		},
		func(tb testing.TB, ic in.TestContext) {
			transaction = w.GetTransaction(filepath.Join(ic.WorkingDir(), "one"))
			err := transaction.Begin()
			assert.NoError(t, err)
		},
		updateModFile("one"),
		func(tb testing.TB, ic in.TestContext) {
			err := transaction.ModifiedFiles(filepath.Join(ic.WorkingDir(), "one", "go.mod"))
			assert.NoError(t, err)
		},
		func(tb testing.TB, ic in.TestContext) {
			waitForEvents(t, events, []WatchEvent{})
			topic.Close()
		},
	)
}

func TestWatchWithBuildAndUserModifyingFiles(t *testing.T) {
	var events chan WatchEvent
	var topic *pubsub.Topic[WatchEvent]
	var transaction ModifyFilesTransaction
	w := NewWatcher("**/*.go", "go.mod", "go.sum")

	in.Run(t,
		func(tb testing.TB, ic in.TestContext) {
			events, topic = startWatching(ic, t, w, ic.WorkingDir())
		},

		in.FtlNew("go", "one"),
		func(tb testing.TB, ic in.TestContext) {
			waitForEventsWhileIgnoringEvents(tb, events, []WatchEvent{
				WatchEventModuleAdded{Config: loadModule(t, ic.WorkingDir(), "one")},
			}, []WatchEvent{
				WatchEventModuleChanged{Config: loadModule(t, ic.WorkingDir(), "one")},
			})
		},
		// Change a file in a module, within a transaction
		func(tb testing.TB, ic in.TestContext) {
			transaction = w.GetTransaction(filepath.Join(ic.WorkingDir(), "one"))
			err := transaction.Begin()
			assert.NoError(t, err)
		},
		updateModFile("one"),
		// Change a file in a module, without a transaction (user change)
		in.MoveFile("one", "one.go", "one_.go"),
		func(tb testing.TB, ic in.TestContext) {
			err := transaction.End()
			assert.NoError(t, err)
		},
		func(tb testing.TB, ic in.TestContext) {
			waitForEvents(t, events, []WatchEvent{
				WatchEventModuleChanged{Config: loadModule(t, ic.WorkingDir(), "one")},
			})
			topic.Close()
		},
	)
}

func loadModule(t *testing.T, dir, name string) moduleconfig.UnvalidatedModuleConfig {
	t.Helper()
	config, err := moduleconfig.LoadConfig(filepath.Join(dir, name))
	assert.NoError(t, err)
	return config
}

func startWatching(ctx context.Context, t testing.TB, w *Watcher, dir string) (chan WatchEvent, *pubsub.Topic[WatchEvent]) {
	t.Helper()
	events := make(chan WatchEvent, 128)
	topic, err := w.Watch(ctx, pollFrequency, []string{dir})
	assert.NoError(t, err)
	topic.Subscribe(events)

	return events, topic
}

// waitForEvents waits for the expected events to be received on the events channel.
//
// It always waits for longer than just the expected events to confirm that no other events are received.
// The expected events are matched by keyForEvent.
func waitForEvents(t testing.TB, events chan WatchEvent, expected []WatchEvent) {
	t.Helper()

	waitForEventsWhileIgnoringEvents(t, events, expected, []WatchEvent{})
}

func waitForEventsWhileIgnoringEvents(t testing.TB, events chan WatchEvent, expected []WatchEvent, ignoredEvents []WatchEvent) {
	t.Helper()
	visited := map[string]bool{}
	expectedKeys := []string{}
	for _, event := range expected {
		key := keyForEvent(event)
		visited[key] = false
		expectedKeys = append(expectedKeys, key)
	}
	ignored := map[string]bool{}
	for _, event := range ignoredEvents {
		key := keyForEvent(event)
		ignored[key] = true
	}
	eventCount := 0
	for {
		select {
		case actual := <-events:
			key := keyForEvent(actual)
			_, isIgnored := ignored[key]
			if isIgnored {
				continue
			}
			hasVisited, isExpected := visited[key]
			assert.True(t, isExpected, "unexpected event %v instead of %v", key, expectedKeys)
			assert.False(t, hasVisited, "duplicate event %v", key)
			visited[key] = true

			eventCount++
		case <-time.After(pollFrequency * 5):
			if eventCount == len(expected) {
				return
			}
			t.Fatalf("timed out waiting for events: %v", visited)
		}
	}
}

func keyForEvent(event WatchEvent) string {
	switch event := event.(type) {
	case WatchEventModuleAdded:
		return "added:" + event.Config.Module
	case WatchEventModuleRemoved:
		return "removed:" + event.Config.Module
	case WatchEventModuleChanged:
		return "updated:" + event.Config.Module
	default:
		panic("unknown event type")
	}
}

func updateModFile(module string) in.Action {
	return in.EditFile(module, func(b []byte) []byte {
		return []byte(strings.Replace(string(b), "github.com/block/ftl", "../..", 1))
	}, "go.mod")
}
