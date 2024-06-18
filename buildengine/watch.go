package buildengine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alecthomas/types/pubsub"

	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/maps"
)

// A WatchEvent is an event that occurs when a module is added, removed, or
// changed.
type WatchEvent interface{ watchEvent() }

type WatchEventModuleAdded struct{ Module Module }

func (WatchEventModuleAdded) watchEvent() {}

type WatchEventModuleRemoved struct{ Module Module }

func (WatchEventModuleRemoved) watchEvent() {}

type WatchEventModuleChanged struct {
	Module Module
	Change FileChangeType
	Path   string
	Time   time.Time
}

func (WatchEventModuleChanged) watchEvent() {}

type moduleHashes struct {
	Hashes FileHashes
	Module Module
}

type Watcher struct {
	isWatching bool

	// use mutex whenever accessing / modifying existingModules or moduleTransactions
	mutex              sync.Mutex
	existingModules    map[string]moduleHashes
	moduleTransactions map[string][]*modifyFilesTransaction
}

func NewWatcher() *Watcher {
	svc := &Watcher{
		existingModules:    map[string]moduleHashes{},
		moduleTransactions: map[string][]*modifyFilesTransaction{},
	}

	return svc
}

func (w *Watcher) GetTransaction(moduleDir string) ModifyFilesTransaction {
	return &modifyFilesTransaction{
		watcher:   w,
		moduleDir: moduleDir,
	}
}

// Watch the given directories for new modules, deleted modules, and changes to
// existing modules, publishing a change event for each.
func (w *Watcher) Watch(ctx context.Context, period time.Duration, moduleDirs []string) (*pubsub.Topic[WatchEvent], error) {
	if w.isWatching {
		return nil, fmt.Errorf("file watcher is already watching")
	}
	w.isWatching = true

	logger := log.FromContext(ctx)
	topic := pubsub.New[WatchEvent]()

	go func() {
		wait := topic.Wait()
		for {
			select {
			case <-time.After(period):

			case <-wait:
				return

			case <-ctx.Done():
				_ = topic.Close()
				return
			}

			modules, err := DiscoverModules(ctx, moduleDirs)
			if err != nil {
				logger.Tracef("error discovering modules: %v", err)
				continue
			}

			modulesByDir := maps.FromSlice(modules, func(module Module) (string, Module) {
				return module.Config.Dir, module
			})

			w.mutex.Lock()
			// Trigger events for removed modules.
			for _, existingModule := range w.existingModules {
				if transactions, ok := w.moduleTransactions[existingModule.Module.Config.Dir]; ok && len(transactions) > 0 {
					// Skip modules that currently have transactions
					continue
				}
				existingConfig := existingModule.Module.Config
				if _, haveModule := modulesByDir[existingConfig.Dir]; !haveModule {
					logger.Debugf("removed %q", existingModule.Module.Config.Module)
					topic.Publish(WatchEventModuleRemoved{Module: existingModule.Module})
					delete(w.existingModules, existingConfig.Dir)
				}
			}

			// Compare the modules to the existing modules.
			for _, module := range modulesByDir {
				config := module.Config
				if transactions, ok := w.moduleTransactions[config.Dir]; ok && len(transactions) > 0 {
					// Skip modules that currently have transactions
					continue
				}
				existingModule, haveExistingModule := w.existingModules[config.Dir]
				hashes, err := ComputeFileHashes(module)
				if err != nil {
					logger.Tracef("error computing file hashes for %s: %v", config.Dir, err)
					continue
				}

				if haveExistingModule {
					changeType, path, equal := CompareFileHashes(existingModule.Hashes, hashes)
					if equal {
						continue
					}
					logger.Debugf("changed %q: %c%s", config.Module, changeType, path)
					topic.Publish(WatchEventModuleChanged{Module: existingModule.Module, Change: changeType, Path: path, Time: time.Now()})
					w.existingModules[config.Dir] = moduleHashes{Hashes: hashes, Module: existingModule.Module}
					continue
				}
				logger.Debugf("added %q", config.Module)
				topic.Publish(WatchEventModuleAdded{Module: module})
				w.existingModules[config.Dir] = moduleHashes{Hashes: hashes, Module: module}
			}
			w.mutex.Unlock()
		}
	}()
	return topic, nil
}

// ModifyFilesTransaction allows builds to modify files in a module without triggering a watch event.
// This helps us avoid infinite loops with builds changing files, and those changes triggering new builds.as a no-op
type ModifyFilesTransaction interface {
	Begin() error
	ModifiedFiles(paths ...string) error
	End() error
}

// Implementation of ModifyFilesTransaction protocol
type modifyFilesTransaction struct {
	watcher   *Watcher
	moduleDir string
	isActive  bool
}

var _ ModifyFilesTransaction = (*modifyFilesTransaction)(nil)
var _ compile.ModifyFilesTransaction = (*modifyFilesTransaction)(nil)

func (t *modifyFilesTransaction) Begin() error {
	if t.isActive {
		return fmt.Errorf("transaction is already active")
	}
	t.isActive = true

	t.watcher.mutex.Lock()
	defer t.watcher.mutex.Unlock()

	t.watcher.moduleTransactions[t.moduleDir] = append(t.watcher.moduleTransactions[t.moduleDir], t)

	return nil
}

func (t *modifyFilesTransaction) End() error {
	if !t.isActive {
		return fmt.Errorf("transaction is not active")
	}

	t.watcher.mutex.Lock()
	defer t.watcher.mutex.Unlock()

	for idx, transaction := range t.watcher.moduleTransactions[t.moduleDir] {
		if transaction != t {
			continue
		}
		t.isActive = false
		t.watcher.moduleTransactions[t.moduleDir] = append(t.watcher.moduleTransactions[t.moduleDir][:idx], t.watcher.moduleTransactions[t.moduleDir][idx+1:]...)
		return nil
	}
	return fmt.Errorf("could not end transaction because it was not found")
}

func (t *modifyFilesTransaction) ModifiedFiles(paths ...string) error {
	if !t.isActive {
		return fmt.Errorf("can not modify file because transaction is not active: %v", paths)
	}

	t.watcher.mutex.Lock()
	defer t.watcher.mutex.Unlock()

	moduleHashes, ok := t.watcher.existingModules[t.moduleDir]
	if !ok {
		// skip updating hashes because we have not discovered this module yet
		return nil
	}

	for _, path := range paths {
		hash, matched, err := ComputeFileHash(moduleHashes.Module.Config.Dir, path, moduleHashes.Module.Config.Watch)
		if err != nil {
			return err
		}
		if !matched {
			continue
		}

		moduleHashes.Hashes[path] = hash
	}
	t.watcher.existingModules[t.moduleDir] = moduleHashes

	return nil
}
