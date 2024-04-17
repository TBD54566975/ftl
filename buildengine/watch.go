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
	Time    time.Time
}

func (WatchEventProjectChanged) watchEvent() {}

// Watch the given directories for new projects, deleted projects, and changes to
// existing projects, publishing a change event for each.
func Watch(ctx context.Context, period time.Duration, moduleDirs []string, externalLibDirs []string) *pubsub.Topic[WatchEvent] {
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

			projects, _ := DiscoverProjects(ctx, moduleDirs, externalLibDirs, false)

			projectsByDir := maps.FromSlice(projects, func(project Project) (string, Project) {
				return project.Config().Dir, project
			})

			// Trigger events for removed projects.
			for _, existingProject := range existingProjects {
				existingConfig := existingProject.Project.Config()
				if _, haveProject := projectsByDir[existingConfig.Dir]; !haveProject {
					logger.Debugf("removed %s %q", existingProject.Project.TypeString(), existingProject.Project.Config().Key)
					topic.Publish(WatchEventProjectRemoved{Project: existingProject.Project})
					delete(existingProjects, existingConfig.Dir)
				}
			}

			// Compare the projects to the existing projects.
			for _, project := range projectsByDir {
				config := project.Config()
				existingProject, haveExistingProject := existingProjects[config.Dir]
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
					existingProjects[config.Dir] = projectHashes{Hashes: hashes, Project: existingProject.Project}
					continue
				}
				logger.Debugf("added %s %q", project.TypeString(), project.Config().Key)
				topic.Publish(WatchEventProjectAdded{Project: project})
				existingProjects[config.Dir] = projectHashes{Hashes: hashes, Project: project}
			}
		}
	}()
	return topic
}
