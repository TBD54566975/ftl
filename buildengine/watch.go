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

// A WatchEvent is an event that occurs when a project is added, removed, or
// changed.
type WatchEvent interface{ watchEvent() }

type WatchEventProjectAdded struct{ Project Project }

func (WatchEventProjectAdded) watchEvent() {}

type WatchEventProjectRemoved struct{ Project Project }

func (WatchEventProjectRemoved) watchEvent() {}

type WatchEventProjectChanged struct {
	Project Project
	Change  FileChangeType
	Path    string
	Time    time.Time
}

func (WatchEventProjectChanged) watchEvent() {}

type projectHashes struct {
	Hashes  FileHashes
	Project Project
}

type Watcher struct {
	isWatching bool

	// use mutex whenever accessing / modifying existingProjects or moduleTransactions
	mutex              sync.Mutex
	existingProjects   map[string]projectHashes
	moduleTransactions map[string][]*modifyFilesTransaction
}

func NewWatcher() *Watcher {
	svc := &Watcher{
		existingProjects:   map[string]projectHashes{},
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

// Watch the given directories for new projects, deleted projects, and changes to
// existing projects, publishing a change event for each.
func (w *Watcher) Watch(ctx context.Context, period time.Duration, moduleDirs []string, externalLibDirs []string) (*pubsub.Topic[WatchEvent], error) {
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

			projects, _ := DiscoverProjects(ctx, moduleDirs, externalLibDirs, false)

			projectsByDir := maps.FromSlice(projects, func(project Project) (string, Project) {
				return project.Config().Dir, project
			})

			w.mutex.Lock()
			// Trigger events for removed projects.
			for _, existingProject := range w.existingProjects {
				if transactions, ok := w.moduleTransactions[existingProject.Project.Config().Dir]; ok && len(transactions) > 0 {
					// Skip projects that currently have transactions
					continue
				}
				existingConfig := existingProject.Project.Config()
				if _, haveProject := projectsByDir[existingConfig.Dir]; !haveProject {
					logger.Debugf("removed %s %q", existingProject.Project.TypeString(), existingProject.Project.Config().Key)
					topic.Publish(WatchEventProjectRemoved{Project: existingProject.Project})
					delete(w.existingProjects, existingConfig.Dir)
				}
			}

			// Compare the projects to the existing projects.
			for _, project := range projectsByDir {
				if transactions, ok := w.moduleTransactions[project.Config().Dir]; ok && len(transactions) > 0 {
					// Skip projects that currently have transactions
					continue
				}
				config := project.Config()
				existingProject, haveExistingProject := w.existingProjects[config.Dir]
				hashes, err := ComputeFileHashes(project)
				if err != nil {
					logger.Tracef("error computing file hashes for %s: %v", config.Dir, err)
					continue
				}

				if haveExistingProject {
					changeType, path, equal := CompareFileHashes(existingProject.Hashes, hashes)
					if equal {
						continue
					}
					logger.Debugf("changed %s %q: %c%s", project.TypeString(), project.Config().Key, changeType, path)
					topic.Publish(WatchEventProjectChanged{Project: existingProject.Project, Change: changeType, Path: path, Time: time.Now()})
					w.existingProjects[config.Dir] = projectHashes{Hashes: hashes, Project: existingProject.Project}
					continue
				}
				logger.Debugf("added %s %q", project.TypeString(), project.Config().Key)
				topic.Publish(WatchEventProjectAdded{Project: project})
				w.existingProjects[config.Dir] = projectHashes{Hashes: hashes, Project: project}
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

	projectHashes, ok := t.watcher.existingProjects[t.moduleDir]
	if !ok {
		// skip updating hashes because we have not discovered this project yet
		return nil
	}

	for _, path := range paths {
		hash, matched, err := ComputeFileHash(projectHashes.Project, path)
		if err != nil {
			return err
		}
		if !matched {
			continue
		}

		projectHashes.Hashes[path] = hash
	}
	t.watcher.existingProjects[t.moduleDir] = projectHashes

	return nil
}
