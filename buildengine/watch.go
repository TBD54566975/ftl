package buildengine

import (
	"context"
	"time"

	"github.com/alecthomas/types/pubsub"

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
}

func (WatchEventProjectChanged) watchEvent() {}

// Watch the given directories for new projects, deleted projects, and changes to
// existing projects, publishing a change event for each.
func Watch(ctx context.Context, period time.Duration, dirs []string, externalLibDirs []string) *pubsub.Topic[WatchEvent] {
	logger := log.FromContext(ctx)
	topic := pubsub.New[WatchEvent]()
	go func() {
		type projectHashes struct {
			Hashes  FileHashes
			Project Project
		}
		existingProjects := map[string]projectHashes{}
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

			projects := projectsIn(logger, dirs, externalLibDirs)

			projectsByDir := maps.FromSlice(projects, func(project Project) (string, Project) {
				return project.Dir(), project
			})

			// Trigger events for removed projects.
			for _, existingProject := range existingProjects {
				if _, haveProject := projectsByDir[existingProject.Project.Dir()]; !haveProject {
					if _, ok := existingProject.Project.(Module); ok {
						logger.Debugf("%s %s removed: %s", existingProject.Project.Kind(), existingProject.Project.Key(), existingProject.Project.Dir())
					} else {
						logger.Debugf("%s removed: %s", existingProject.Project.Kind(), existingProject.Project.Dir())
					}
					topic.Publish(WatchEventProjectRemoved{Project: existingProject.Project})
					delete(existingProjects, existingProject.Project.Dir())
				}
			}

			// Compare the projects to the existing projects.
			for _, project := range projectsByDir {
				existingProject, haveExistingProject := existingProjects[project.Dir()]
				hashes, err := ComputeFileHashes(project)
				if err != nil {
					logger.Tracef("error computing file hashes for %s: %v", project.Dir(), err)
					continue
				}

				if haveExistingProject {
					changeType, path, equal := CompareFileHashes(existingProject.Hashes, hashes)
					if equal {
						continue
					}
					logger.Debugf("%s %s changed: %c%s", project.Kind(), project.Key(), changeType, path)
					topic.Publish(WatchEventProjectChanged{Project: existingProject.Project, Change: changeType, Path: path})
					existingProjects[project.Dir()] = projectHashes{Hashes: hashes, Project: existingProject.Project}
					continue
				}
				if _, ok := project.(Module); ok {
					logger.Debugf("%s %s added: %s", project.Kind(), project.Key(), project.Dir())
				} else {
					logger.Debugf("%s added: %s", project.Kind(), project.Dir())
				}

				topic.Publish(WatchEventProjectAdded{Project: project})
				existingProjects[project.Dir()] = projectHashes{Hashes: hashes, Project: project}
			}
		}
	}()
	return topic
}

func projectsIn(logger *log.Logger, dirs []string, externalLibDirs []string) []Project {
	out := []Project{}

	modules, err := DiscoverModules(context.Background(), dirs...)
	if err != nil {
		logger.Tracef("error discovering modules: %v", err)
	} else {
		for _, module := range modules {
			out = append(out, Project(&module))
		}
	}

	for _, dir := range externalLibDirs {
		if lib, err := LoadExternalLibrary(dir); err == nil {
			out = append(out, Project(&lib))
		}
	}

	return out
}
