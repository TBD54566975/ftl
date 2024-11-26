package schemaeventsource

import (
	"context"
	"fmt"
	"slices"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/reflect"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

// Event represents a change in the schema.
//
//sumtype:decl
type Event interface {
	// More returns true if there are more changes to come as part of the initial sync.
	More() bool
	// Schema is the READ-ONLY full schema after this event was applied.
	Schema() *schema.Schema
	change()
}

// EventRemove represents that a deployment (or module) was removed.
type EventRemove struct {
	// None for builtin modules.
	Deployment optional.Option[model.DeploymentKey]
	Module     *schema.Module
	// True if the underlying module was deleted in addition to the deployment itself.
	Deleted bool

	schema *schema.Schema
	more   bool
}

func (c EventRemove) change()                {}
func (c EventRemove) More() bool             { return c.more }
func (c EventRemove) Schema() *schema.Schema { return c.schema }

// EventUpsert represents that a module has been added or updated in the schema.
type EventUpsert struct {
	// schema is the READ-ONLY full schema after this event was applied.
	schema *schema.Schema

	// None for builtin modules.
	Deployment optional.Option[model.DeploymentKey]
	Module     *schema.Module

	more bool
}

func (c EventUpsert) change()                {}
func (c EventUpsert) More() bool             { return c.more }
func (c EventUpsert) Schema() *schema.Schema { return c.schema }

// NewUnattached creates a new EventSource that is not attached to a SchemaService.
func NewUnattached() EventSource {
	return EventSource{
		events: make(chan Event, 64),
		view:   atomic.New[*schema.Schema](&schema.Schema{}),
	}
}

// EventSource represents a stream of schema events and the materialised view of those events.
type EventSource struct {
	events chan Event
	view   *atomic.Value[*schema.Schema]
}

// Events is a stream of schema change events. "View" will be updated with these changes prior to being sent on this
// channel.
func (e EventSource) Events() <-chan Event { return e.events }

// View is the materialised view of the schema from "Events".
func (e EventSource) View() *schema.Schema { return e.view.Load() }

// Publish an event to the EventSource.
//
// This will update the materialised view and send the event on the "Events" channel. The event will be updated with the
// materialised view.
//
// This is mostly useful in conjunction with NewUnattached, for testing.
func (e EventSource) Publish(event Event) {
	clone := reflect.DeepCopy(e.View())
	switch event := event.(type) {
	case EventRemove:
		if event.Deleted {
			clone.Modules = slices.DeleteFunc(clone.Modules, func(m *schema.Module) bool { return m.Name == event.Module.Name })
		}
		event.schema = clone
		e.view.Store(clone)
		e.events <- event

	case EventUpsert:
		if i := slices.IndexFunc(clone.Modules, func(m *schema.Module) bool { return m.Name == event.Module.Name }); i != -1 {
			clone.Modules[i] = event.Module
		} else {
			clone.Modules = append(clone.Modules, event.Module)
		}
		event.schema = clone
		e.view.Store(clone)
		e.events <- event
	}
}

// New creates a new EventSource that pulls schema changes from the SchemaService into an event channel and a
// materialised view (ie. [schema.Schema]).
//
// The sync will terminate when the context is cancelled.
func New(ctx context.Context, client ftlv1connect.SchemaServiceClient) EventSource {
	logger := log.FromContext(ctx).Scope("schema-sync")
	out := NewUnattached()
	more := true
	logger.Debugf("Starting schema pull")
	go rpc.RetryStreamingServerStream(ctx, "schema-sync", backoff.Backoff{}, &ftlv1.PullSchemaRequest{}, client.PullSchema, func(_ context.Context, resp *ftlv1.PullSchemaResponse) error {
		sch, err := schema.ModuleFromProto(resp.Schema)
		if err != nil {
			return fmt.Errorf("schema-sync: failed to decode module schema: %w", err)
		}
		var someDeploymentKey optional.Option[model.DeploymentKey]
		if resp.DeploymentKey != nil {
			deploymentKey, err := model.ParseDeploymentKey(resp.GetDeploymentKey())
			if err != nil {
				return fmt.Errorf("schema-sync: invalid deployment key %q: %w", resp.GetDeploymentKey(), err)
			}
			someDeploymentKey = optional.Some(deploymentKey)
		}
		// resp.More can become true again if the streaming client reconnects, but we don't want downstream to have to
		// care about a new initial sync restarting.
		more = more && resp.More
		switch resp.ChangeType {
		case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
			logger.Debugf("Module %s removed", sch.Name)
			event := EventRemove{
				Deployment: someDeploymentKey,
				Module:     sch,
				Deleted:    resp.ModuleRemoved,
				more:       more,
			}
			out.Publish(event)

		case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:
			logger.Debugf("Module %s upserted", sch.Name)
			event := EventUpsert{
				Deployment: someDeploymentKey,
				Module:     sch,
				more:       more,
			}
			out.Publish(event)

		default:
			return fmt.Errorf("schema-sync: unknown change type %q", resp.ChangeType)
		}
		return nil
	}, rpc.AlwaysRetry())
	return out
}
