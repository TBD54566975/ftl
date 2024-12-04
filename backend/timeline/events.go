package timeline

import (
	"fmt"
	"slices"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
)

func filter(event *timelinepb.Event, depKey string, eventTypes []timelinepb.EventType) bool {
	if !slices.Contains(eventTypes, eventType(event)) {
		return false
	}
	if depKey != "" && depKey != deploymentKey(event) {
		return false
	}
	return true
}

func eventType(event *timelinepb.Event) timelinepb.EventType {
	switch event.Entry.(type) {
	case *timelinepb.Event_Log:
		return timelinepb.EventType_EVENT_TYPE_LOG
	case *timelinepb.Event_Call:
		return timelinepb.EventType_EVENT_TYPE_CALL
	case *timelinepb.Event_DeploymentCreated:
		return timelinepb.EventType_EVENT_TYPE_DEPLOYMENT_CREATED
	case *timelinepb.Event_DeploymentUpdated:
		return timelinepb.EventType_EVENT_TYPE_DEPLOYMENT_UPDATED
	case *timelinepb.Event_Ingress:
		return timelinepb.EventType_EVENT_TYPE_INGRESS
	case *timelinepb.Event_CronScheduled:
		return timelinepb.EventType_EVENT_TYPE_CRON_SCHEDULED
	case *timelinepb.Event_AsyncExecute:
		return timelinepb.EventType_EVENT_TYPE_ASYNC_EXECUTE
	case *timelinepb.Event_PubsubPublish:
		return timelinepb.EventType_EVENT_TYPE_PUBSUB_PUBLISH
	case *timelinepb.Event_PubsubConsume:
		return timelinepb.EventType_EVENT_TYPE_PUBSUB_CONSUME
	default:
		panic(fmt.Sprintf("unexpected event type: %T", event.Entry))
	}
}

func deploymentKey(event *timelinepb.Event) string {
	switch entry := event.Entry.(type) {
	case *timelinepb.Event_Log:
		return entry.Log.DeploymentKey
	case *timelinepb.Event_Call:
		return entry.Call.DeploymentKey
	case *timelinepb.Event_DeploymentCreated:
		return entry.DeploymentCreated.Key
	case *timelinepb.Event_DeploymentUpdated:
		return entry.DeploymentUpdated.Key
	case *timelinepb.Event_Ingress:
		return entry.Ingress.DeploymentKey
	case *timelinepb.Event_CronScheduled:
		return entry.CronScheduled.DeploymentKey
	case *timelinepb.Event_AsyncExecute:
		return entry.AsyncExecute.DeploymentKey
	case *timelinepb.Event_PubsubPublish:
		return entry.PubsubPublish.DeploymentKey
	case *timelinepb.Event_PubsubConsume:
		return entry.PubsubConsume.DeploymentKey
	default:
		panic(fmt.Sprintf("unexpected event type: %T", event.Entry))
	}
}
