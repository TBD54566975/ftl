package buildengine

import (
	"context"
	"time"

	"github.com/alecthomas/types/pubsub"

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
}

func (WatchEventModuleChanged) watchEvent() {}

// Watch the given directories for new modules, deleted modules, and changes to
// existing modules, publishing a change event for each.
func Watch(ctx context.Context, period time.Duration, dirs []string, externalLibDirs []string) *pubsub.Topic[WatchEvent] {
	logger := log.FromContext(ctx)
	topic := pubsub.New[WatchEvent]()
	go func() {
		type moduleHashes struct {
			Hashes FileHashes
			Module Module
		}
		existingModules := map[string]moduleHashes{}
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

			// Find all modules in the given directories.
			modules, err := DiscoverModules(ctx, dirs...)
			if err != nil {
				logger.Tracef("error discovering modules: %v", err)
				continue
			}
			for _, dir := range externalLibDirs {
				if module, err := LoadExternalLibrary(ctx, dir); err == nil {
					modules = append(modules, module)
				}
			}
			modulesByDir := maps.FromSlice(modules, func(module Module) (string, Module) {
				return module.Dir(), module
			})

			// Trigger events for removed modules.
			for _, existingModule := range existingModules {
				if _, haveModule := modulesByDir[existingModule.Module.Dir()]; !haveModule {
					if _, ok := existingModule.Module.ModuleConfig(); ok {
						logger.Debugf("%s %s removed: %s", existingModule.Module.Kind(), existingModule.Module.Key(), existingModule.Module.Dir())
					} else {
						logger.Debugf("%s removed: %s", existingModule.Module.Kind(), existingModule.Module.Dir())
					}
					topic.Publish(WatchEventModuleRemoved{Module: existingModule.Module})
					delete(existingModules, existingModule.Module.Dir())
				}
			}

			// Compare the modules to the existing modules.
			for _, module := range modulesByDir {
				existingModule, haveExistingModule := existingModules[module.Dir()]
				hashes, err := ComputeFileHashes(module)
				if err != nil {
					logger.Tracef("error computing file hashes for %s: %v", module.Dir(), err)
					continue
				}

				if haveExistingModule {
					changeType, path, equal := CompareFileHashes(existingModule.Hashes, hashes)
					if equal {
						continue
					}
					logger.Debugf("%s %s changed: %c%s", module.Kind(), module.Key(), changeType, path)
					topic.Publish(WatchEventModuleChanged{Module: existingModule.Module, Change: changeType, Path: path})
					existingModules[module.Dir()] = moduleHashes{Hashes: hashes, Module: existingModule.Module}
					continue
				}
				if _, ok := module.ModuleConfig(); ok {
					logger.Debugf("%s %s added: %s", module.Kind(), module.Key(), module.Dir())
				} else {
					logger.Debugf("%s added: %s", module.Kind(), module.Dir())
				}

				topic.Publish(WatchEventModuleAdded{Module: module})
				existingModules[module.Dir()] = moduleHashes{Hashes: hashes, Module: module}
			}
		}
	}()
	return topic
}
