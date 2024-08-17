package buildengine_test

import (
	"context"
	"os"
	"os/exec" //nolint:depguard
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/pubsub"

	. "github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
)

const pollFrequency = time.Millisecond * 500

func TestWatch(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	dir := t.TempDir()

	w := NewWatcher()
	events, topic := startWatching(ctx, t, w, dir)

	waitForEvents(t, events, []WatchEvent{})

	// Initiate two modules
	err := gitInit(dir)
	assert.NoError(t, err)
	err = ftl("init", "test", dir)
	assert.NoError(t, err)
	err = ftl("new", "go", dir, "one")
	assert.NoError(t, err)
	err = ftl("new", "go", dir, "two")
	assert.NoError(t, err)

	one := loadModule(t, dir, "one")
	two := loadModule(t, dir, "two")

	waitForEvents(t, events, []WatchEvent{
		WatchEventModuleAdded{Module: one},
		WatchEventModuleAdded{Module: two},
	})

	// Delete a module
	err = os.RemoveAll(filepath.Join(dir, "two"))
	assert.NoError(t, err)

	// Change a module.
	updateModFile(t, filepath.Join(dir, "one"))

	waitForEvents(t, events, []WatchEvent{
		WatchEventModuleChanged{Module: one},
		WatchEventModuleRemoved{Module: two},
	})
	topic.Close()
}

func TestWatchWithBuildModifyingFiles(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	dir := t.TempDir()

	w := NewWatcher()

	// Initiate a module
	err := gitInit(dir)
	assert.NoError(t, err)
	err = ftl("init", "test", dir)
	assert.NoError(t, err)
	err = ftl("new", "go", dir, "one")
	assert.NoError(t, err)

	events, topic := startWatching(ctx, t, w, dir)

	waitForEvents(t, events, []WatchEvent{
		WatchEventModuleAdded{Module: loadModule(t, dir, "one")},
	})

	// Change a file in a module, within a transaction
	transaction := w.GetTransaction(filepath.Join(dir, "one"))
	err = transaction.Begin()
	assert.NoError(t, err)
	updateModFile(t, filepath.Join(dir, "one"))
	err = transaction.ModifiedFiles(filepath.Join(dir, "one", "go.mod"))
	assert.NoError(t, err)

	err = transaction.End()
	assert.NoError(t, err)

	waitForEvents(t, events, []WatchEvent{})
	topic.Close()
}

func TestWatchWithBuildAndUserModifyingFiles(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	dir := t.TempDir()

	// Initiate a module
	err := gitInit(dir)
	assert.NoError(t, err)
	err = ftl("init", "test", dir)
	assert.NoError(t, err)
	err = ftl("new", "go", dir, "one")
	assert.NoError(t, err)

	one := loadModule(t, dir, "one")

	w := NewWatcher()
	events, topic := startWatching(ctx, t, w, dir)

	waitForEvents(t, events, []WatchEvent{
		WatchEventModuleAdded{Module: one},
	})

	// Change a file in a module, within a transaction
	transaction := w.GetTransaction(filepath.Join(dir, "one"))
	err = transaction.Begin()
	assert.NoError(t, err)

	updateModFile(t, filepath.Join(dir, "one"))

	// Change a file in a module, without a transaction (user change)
	cmd := exec.Command("mv", "one.go", "one_.go")
	cmd.Dir = filepath.Join(dir, "one")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	assert.NoError(t, err)

	err = transaction.End()
	assert.NoError(t, err)

	waitForEvents(t, events, []WatchEvent{
		WatchEventModuleChanged{Module: one},
	})
	topic.Close()
}

func loadModule(t *testing.T, dir, name string) Module {
	t.Helper()
	config, err := moduleconfig.LoadModuleConfig(filepath.Join(dir, name))
	assert.NoError(t, err)
	return Module{
		Config: config,
	}
}

func startWatching(ctx context.Context, t *testing.T, w *Watcher, dir string) (chan WatchEvent, *pubsub.Topic[WatchEvent]) {
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
func waitForEvents(t *testing.T, events chan WatchEvent, expected []WatchEvent) {
	t.Helper()
	visited := map[string]bool{}
	expectedKeys := []string{}
	for _, event := range expected {
		key := keyForEvent(event)
		visited[key] = false
		expectedKeys = append(expectedKeys, key)
	}
	eventCount := 0
	for {
		select {
		case actual := <-events:
			key := keyForEvent(actual)
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
		return "added:" + event.Module.Config.Module
	case WatchEventModuleRemoved:
		return "removed:" + event.Module.Config.Module
	case WatchEventModuleChanged:
		return "updated:" + event.Module.Config.Module
	default:
		panic("unknown event type")
	}
}

func gitInit(dir string) error {
	cmd := exec.Command("git", "init", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ftl(args ...string) error {
	cmd := exec.Command("ftl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func updateModFile(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("go", "mod", "edit", "-replace=github.com/TBD54566975/ftl=../..")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	assert.NoError(t, err)
}
