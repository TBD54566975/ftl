package buildengine

import (
	"context"
	"time"

	"github.com/alecthomas/types/pubsub"

	"github.com/TBD54566975/ftl/common/moduleconfig"
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
func Watch(ctx context.Context, period time.Duration, dirs ...string) *pubsub.Topic[WatchEvent] {
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
			moduleConfigs, err := DiscoverModules(dirs...)
			if err != nil {
				logger.Tracef("error discovering modules: %v", err)
				continue
			}
			moduleConfigsByDir := maps.FromSlice(moduleConfigs, func(config moduleconfig.ModuleConfig) (string, moduleconfig.ModuleConfig) {
				return config.Module, config
			})

			// Trigger events for removed modules.
			for _, existingModule := range existingModules {
				if _, haveModule := moduleConfigsByDir[existingModule.Module.Module]; !haveModule {
					logger.Debugf("module %s removed: %s", existingModule.Module.Module, existingModule.Module.Dir)
					topic.Publish(WatchEventModuleRemoved{Module: existingModule.Module})
					delete(existingModules, existingModule.Module.Module)
				}
			}

			// Compare the modules to the existing modules.
			for _, config := range moduleConfigs {
				existingModule, haveExistingModule := existingModules[config.Dir]
				hashes, err := ComputeFileHashes(config)
				if err != nil {
					logger.Tracef("error computing file hashes for %s: %v", config.Dir, err)
					continue
				}

				if haveExistingModule {
					changeType, path, equal := CompareFileHashes(existingModule.Hashes, hashes)
					if equal {
						continue
					}
					logger.Debugf("module %s changed: %c%s", existingModule.Module.Module, changeType, path)
					topic.Publish(WatchEventModuleChanged{Module: existingModule.Module, Change: changeType, Path: path})
					existingModules[config.Dir] = moduleHashes{Hashes: hashes, Module: existingModule.Module}
					continue
				}

				module, err := UpdateDependencies(config)
				if err != nil {
					continue
				}
				logger.Debugf("module %s added: %s", module.Module, module.Dir)
				topic.Publish(WatchEventModuleAdded{Module: module})
				existingModules[config.Dir] = moduleHashes{Hashes: hashes, Module: module}
			}
		}
	}()
	return topic
}
