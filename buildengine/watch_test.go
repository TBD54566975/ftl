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

	s := NewWatcher()

	// Start the watch
	events := make(chan WatchEvent, 128)
	topic, err := s.Watch(ctx, time.Millisecond*200, []string{dir}, nil)
	assert.NoError(t, err)
	topic.Subscribe(events)

	// Initiate a bunch of changes.
	err = ftl("init", "go", dir, "one")
	assert.NoError(t, err)
	err = ftl("init", "go", dir, "two")
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 750)

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
	topic.Close()

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

func TestWatchWithBuildModifyingFiles(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	dir := t.TempDir()

	s := NewWatcher()

	// Initiate a module
	err := ftl("init", "go", dir, "one")
	assert.NoError(t, err)

	// Start the watch
	events := make(chan WatchEvent, 128)
	topic, err := s.Watch(ctx, time.Millisecond*500, []string{dir}, nil)
	assert.NoError(t, err)
	topic.Subscribe(events)

	time.Sleep(time.Millisecond * 750)

	// Change a file in a module, within a transaction
	transaction := s.GetTransaction(filepath.Join(dir, "one"))
	err = transaction.Begin()
	assert.NoError(t, err)
	cmd := exec.Command("go", "mod", "edit", "-replace=github.com/TBD54566975/ftl=..")
	cmd.Dir = filepath.Join(dir, "one")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	assert.NoError(t, err)
	err = transaction.ModifiedFiles(filepath.Join(dir, "one", "go.mod"))
	assert.NoError(t, err)
	err = transaction.End()
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 500)
	topic.Close()

	allEvents := []WatchEvent{}
	for event := range events {
		allEvents = append(allEvents, event)
	}
	for _, event := range allEvents {
		event, wasAdded := event.(WatchEventProjectAdded)
		assert.True(t, wasAdded, "expected only project added events, got %v", event)
	}
}

func TestWatchWithBuildAndUserModifyingFiles(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	dir := t.TempDir()

	s := NewWatcher()

	// Initiate a module
	err := ftl("init", "go", dir, "one")
	assert.NoError(t, err)

	// Start the watch
	events := make(chan WatchEvent, 128)
	topic, err := s.Watch(ctx, time.Millisecond*500, []string{dir}, nil)
	assert.NoError(t, err)
	topic.Subscribe(events)

	time.Sleep(time.Millisecond * 750)

	// Change a file in a module, within a transaction
	transaction := s.GetTransaction(filepath.Join(dir, "one"))
	err = transaction.Begin()
	assert.NoError(t, err)
	cmd := exec.Command("go", "mod", "edit", "-replace=github.com/TBD54566975/ftl=..")
	cmd.Dir = filepath.Join(dir, "one")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	assert.NoError(t, err)
	err = transaction.ModifiedFiles(filepath.Join(dir, "one", "go.mod"))
	assert.NoError(t, err)
	err = transaction.End()
	assert.NoError(t, err)

	// Change a file in a module, without a transaction (user change)
	cmd = exec.Command("mv", "one.go", "one_.go")
	cmd.Dir = filepath.Join(dir, "one")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 500)
	topic.Close()

	allEvents := []WatchEvent{}
	for event := range events {
		allEvents = append(allEvents, event)
	}
	foundChange := false
	for _, event := range allEvents {
		switch event := event.(type) {
		case WatchEventProjectAdded:
			// expected
		case WatchEventProjectRemoved:
			assert.False(t, true, "unexpected project removed event")
		case WatchEventProjectChanged:
			if event.Project.Config().Key == "one" {
				foundChange = true
			}
		}
	}
	assert.True(t, foundChange, "expected project changed event")
}

func ftl(args ...string) error {
	cmd := exec.Command("ftl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
