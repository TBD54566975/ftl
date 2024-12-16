package buildengine

import (
	"context"

	"github.com/alecthomas/types/pubsub"

	"github.com/block/ftl/internal/terminal"
)

func updateTerminalWithEngineEvents(ctx context.Context, topic *pubsub.Topic[EngineEvent]) {
	events := make(chan EngineEvent, 64)
	topic.Subscribe(events)

	go func() {
		defer topic.Unsubscribe(events)
		for {
			select {
			case event := <-events:
				switch event := event.(type) {
				case EngineStarted:
				case EngineEnded:

				case ModuleAdded:
					terminal.UpdateModuleState(ctx, event.Module, terminal.BuildStateWaiting)
				case ModuleRemoved:
					terminal.UpdateModuleState(ctx, event.Module, terminal.BuildStateTerminated)

				case ModuleBuildWaiting:
					terminal.UpdateModuleState(ctx, event.Config.Module, terminal.BuildStateWaiting)
				case ModuleBuildStarted:
					terminal.UpdateModuleState(ctx, event.Config.Module, terminal.BuildStateBuilding)
				case ModuleBuildSuccess:
					terminal.UpdateModuleState(ctx, event.Config.Module, terminal.BuildStateBuilt)
				case ModuleBuildFailed:
					terminal.UpdateModuleState(ctx, event.Config.Module, terminal.BuildStateFailed)

				case ModuleDeployStarted:
					terminal.UpdateModuleState(ctx, event.Module, terminal.BuildStateDeploying)
				case ModuleDeploySuccess:
					terminal.UpdateModuleState(ctx, event.Module, terminal.BuildStateDeployed)
				case ModuleDeployFailed:
					terminal.UpdateModuleState(ctx, event.Module, terminal.BuildStateFailed)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
