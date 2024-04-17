package buildengine_test

import (
	"context"
	"os"
	"os/exec" //nolint:depguard
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	. "github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestWatch(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	dir := t.TempDir()

	// Start the watch
	events := make(chan WatchEvent, 128)
	watch := Watch(ctx, time.Millisecond*200, []string{dir}, nil)
	watch.Subscribe(events)

	// Initiate a bunch of changes.
	err := ftl("init", "go", dir, "one")
	assert.NoError(t, err)
	err = ftl("init", "go", dir, "two")
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	// Delete a module
	err = os.RemoveAll(filepath.Join(dir, "two"))
	assert.NoError(t, err)

	// Change a module.
	cmd := exec.Command("go", "mod", "edit", "-replace=github.com/TBD54566975/ftl=..")
	cmd.Dir = filepath.Join(dir, "one")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 500)
	watch.Close()

	allEvents := []WatchEvent{}
	for event := range events {
		allEvents = append(allEvents, event)
	}

	assert.True(t, len(allEvents) >= 4, "expected at least 4 events, got %d", len(allEvents))

	// Check we've got at least the events we expect.
	found := 0
	for _, event := range allEvents {
		switch event := event.(type) {
		case WatchEventProjectAdded:
			if event.Project.Config().Key == "one" || event.Project.Config().Key == "two" {
				found++
			}

		case WatchEventProjectRemoved:
			if event.Project.Config().Key == "two" {
				found++
			}

		case WatchEventProjectChanged:
			if event.Project.Config().Key == "one" {
				found++
			}
		}
	}
	assert.True(t, found >= 4, "expected at least 4 events, got %d", found)
}

func ftl(args ...string) error {
	cmd := exec.Command("ftl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
