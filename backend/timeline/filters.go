package timeline

import (
	"fmt"
	"reflect"
	"slices"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	islices "github.com/TBD54566975/ftl/common/slices"
)

type TimelineFilter func(event *timelinepb.Event) bool

func FilterLogLevel(f *timelinepb.GetTimelineRequest_LogLevelFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		log, ok := event.Entry.(*timelinepb.Event_Log)
		if !ok {
			// allow non-log events to pass through
			return true
		}
		return int(log.Log.LogLevel) >= int(f.LogLevel)
	}
}

// FilterCall filters call events between the given modules.
//
// Takes a list of filters, with each call event needing to match at least one filter.
func FilterCall(filters []*timelinepb.GetTimelineRequest_CallFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		call, ok := event.Entry.(*timelinepb.Event_Call)
		if !ok {
			// Allow non-call events to pass through.
			return true
		}
		// Allow event if any event filter matches.
		_, ok = islices.Find(filters, func(f *timelinepb.GetTimelineRequest_CallFilter) bool {
			if call.Call.DestinationVerbRef.Module != f.DestModule {
				return false
			}
			if f.DestVerb != nil && call.Call.DestinationVerbRef.Name != *f.DestVerb {
				return false
			}
			if f.SourceModule != nil && call.Call.SourceVerbRef.Module != *f.SourceModule {
				return false
			}
			return true
		})
		return ok
	}
}

func FilterModule(filters []*timelinepb.GetTimelineRequest_ModuleFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		var module, verb string
		switch entry := event.Entry.(type) {
		case *timelinepb.Event_Call:
			module = entry.Call.DestinationVerbRef.Module
			verb = entry.Call.DestinationVerbRef.Name
		case *timelinepb.Event_Ingress:
			module = entry.Ingress.VerbRef.Module
			verb = entry.Ingress.VerbRef.Name
		case *timelinepb.Event_AsyncExecute:
			module = entry.AsyncExecute.VerbRef.Module
			verb = entry.AsyncExecute.VerbRef.Name
		case *timelinepb.Event_PubsubPublish:
			module = entry.PubsubPublish.VerbRef.Module
			verb = entry.PubsubPublish.VerbRef.Name
		case *timelinepb.Event_PubsubConsume:
			module = *entry.PubsubConsume.DestVerbModule
			verb = *entry.PubsubConsume.DestVerbName
		case *timelinepb.Event_Log, *timelinepb.Event_DeploymentCreated, *timelinepb.Event_DeploymentUpdated, *timelinepb.Event_CronScheduled:
			// Block all other event types.
			return false
		default:
			panic(fmt.Sprintf("unexpected event type: %T", event.Entry))
		}
		// Allow event if any module filter matches.
		_, ok := islices.Find(filters, func(f *timelinepb.GetTimelineRequest_ModuleFilter) bool {
			if f.Module != module {
				return false
			}
			if f.Verb != nil && *f.Verb != verb {
				return false
			}
			return true
		})
		return ok
	}
}

func FilterDeployments(filters []*timelinepb.GetTimelineRequest_DeploymentFilter) TimelineFilter {
	deployments := islices.Reduce(filters, []string{}, func(acc []string, f *timelinepb.GetTimelineRequest_DeploymentFilter) []string {
		return append(acc, f.Deployments...)
	})
	return func(event *timelinepb.Event) bool {
		var deployment string
		switch entry := event.Entry.(type) {
		case *timelinepb.Event_Log:
			deployment = entry.Log.DeploymentKey
		case *timelinepb.Event_Call:
			deployment = entry.Call.DeploymentKey
		case *timelinepb.Event_DeploymentCreated:
			deployment = entry.DeploymentCreated.Key
		case *timelinepb.Event_DeploymentUpdated:
			deployment = entry.DeploymentUpdated.Key
		case *timelinepb.Event_Ingress:
			deployment = entry.Ingress.DeploymentKey
		case *timelinepb.Event_CronScheduled:
			deployment = entry.CronScheduled.DeploymentKey
		case *timelinepb.Event_AsyncExecute:
			deployment = entry.AsyncExecute.DeploymentKey
		case *timelinepb.Event_PubsubPublish:
			deployment = entry.PubsubPublish.DeploymentKey
		case *timelinepb.Event_PubsubConsume:
			deployment = entry.PubsubConsume.DeploymentKey
		default:
			panic(fmt.Sprintf("unexpected event type: %T", event.Entry))
		}
		return slices.Contains(deployments, deployment)
	}
}

func FilterRequests(filters []*timelinepb.GetTimelineRequest_RequestFilter) TimelineFilter {
	requests := islices.Reduce(filters, []string{}, func(acc []string, f *timelinepb.GetTimelineRequest_RequestFilter) []string {
		return append(acc, f.Requests...)
	})
	return func(event *timelinepb.Event) bool {
		var request *string
		switch entry := event.Entry.(type) {
		case *timelinepb.Event_Log:
			request = entry.Log.RequestKey
		case *timelinepb.Event_Call:
			request = entry.Call.RequestKey
		case *timelinepb.Event_Ingress:
			request = entry.Ingress.RequestKey
		case *timelinepb.Event_AsyncExecute:
			request = entry.AsyncExecute.RequestKey
		case *timelinepb.Event_PubsubPublish:
			request = entry.PubsubPublish.RequestKey
		case *timelinepb.Event_PubsubConsume:
			request = entry.PubsubConsume.RequestKey
		case *timelinepb.Event_DeploymentCreated, *timelinepb.Event_DeploymentUpdated, *timelinepb.Event_CronScheduled:
		default:
			panic(fmt.Sprintf("unexpected event type: %T", event.Entry))
		}
		if request == nil {
			return false
		}
		return slices.Contains(requests, *request)
	}
}

func FilterTypes(filters ...*timelinepb.GetTimelineRequest_EventTypeFilter) TimelineFilter {
	types := islices.Reduce(filters, []timelinepb.EventType{}, func(acc []timelinepb.EventType, f *timelinepb.GetTimelineRequest_EventTypeFilter) []timelinepb.EventType {
		return append(acc, f.EventTypes...)
	})
	allowsAll := slices.Contains(types, timelinepb.EventType_EVENT_TYPE_UNSPECIFIED)
	return func(event *timelinepb.Event) bool {
		if allowsAll {
			return true
		}
		var eventType timelinepb.EventType
		switch event.Entry.(type) {
		case *timelinepb.Event_Log:
			eventType = timelinepb.EventType_EVENT_TYPE_LOG
		case *timelinepb.Event_Call:
			eventType = timelinepb.EventType_EVENT_TYPE_CALL
		case *timelinepb.Event_DeploymentCreated:
			eventType = timelinepb.EventType_EVENT_TYPE_DEPLOYMENT_CREATED
		case *timelinepb.Event_DeploymentUpdated:
			eventType = timelinepb.EventType_EVENT_TYPE_DEPLOYMENT_UPDATED
		case *timelinepb.Event_Ingress:
			eventType = timelinepb.EventType_EVENT_TYPE_INGRESS
		case *timelinepb.Event_CronScheduled:
			eventType = timelinepb.EventType_EVENT_TYPE_CRON_SCHEDULED
		case *timelinepb.Event_AsyncExecute:
			eventType = timelinepb.EventType_EVENT_TYPE_ASYNC_EXECUTE
		case *timelinepb.Event_PubsubPublish:
			eventType = timelinepb.EventType_EVENT_TYPE_PUBSUB_PUBLISH
		case *timelinepb.Event_PubsubConsume:
			eventType = timelinepb.EventType_EVENT_TYPE_PUBSUB_CONSUME
		default:
			panic(fmt.Sprintf("unexpected event type: %T", event.Entry))
		}
		return slices.Contains(types, eventType)
	}
}

// FilterTimeRange filters events between the given times, inclusive.
func FilterTimeRange(filter *timelinepb.GetTimelineRequest_TimeFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		if filter.NewerThan != nil && event.Timestamp.AsTime().Before(filter.NewerThan.AsTime()) {
			return false
		}
		if filter.OlderThan != nil && event.Timestamp.AsTime().After(filter.OlderThan.AsTime()) {
			return false
		}
		return true
	}
}

// FilterIDRange filters events between the given IDs, inclusive.
func FilterIDRange(filter *timelinepb.GetTimelineRequest_IDFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		if filter.HigherThan != nil && event.Id < *filter.HigherThan {
			return false
		}
		if filter.LowerThan != nil && event.Id > *filter.LowerThan {
			return false
		}
		return true
	}
}

//nolint:maintidx
func filtersFromRequest(req *timelinepb.GetTimelineRequest) (outFilters []TimelineFilter, ascending bool) {
	if req.Order != timelinepb.GetTimelineRequest_ORDER_DESC {
		ascending = true
	}

	// Some filters need to be combined (for OR logic), so we group them by type first.
	reqFiltersByType := map[reflect.Type][]*timelinepb.GetTimelineRequest_Filter{}
	for _, filter := range req.Filters {
		reqFiltersByType[reflect.TypeOf(filter.Filter)] = append(reqFiltersByType[reflect.TypeOf(filter.Filter)], filter)
	}
	if len(reqFiltersByType) == 0 {
		return outFilters, ascending
	}
	for _, filters := range reqFiltersByType {
		switch filters[0].Filter.(type) {
		case *timelinepb.GetTimelineRequest_Filter_LogLevel:
			outFilters = append(outFilters, islices.Map(filters, func(f *timelinepb.GetTimelineRequest_Filter) TimelineFilter {
				return FilterLogLevel(f.GetLogLevel())
			})...)
		case *timelinepb.GetTimelineRequest_Filter_Deployments:
			outFilters = append(outFilters, FilterDeployments(islices.Map(filters, func(f *timelinepb.GetTimelineRequest_Filter) *timelinepb.GetTimelineRequest_DeploymentFilter {
				return f.GetDeployments()
			})))
		case *timelinepb.GetTimelineRequest_Filter_Requests:
			outFilters = append(outFilters, FilterRequests(islices.Map(filters, func(f *timelinepb.GetTimelineRequest_Filter) *timelinepb.GetTimelineRequest_RequestFilter {
				return f.GetRequests()
			})))
		case *timelinepb.GetTimelineRequest_Filter_EventTypes:
			outFilters = append(outFilters, FilterTypes(islices.Map(filters, func(f *timelinepb.GetTimelineRequest_Filter) *timelinepb.GetTimelineRequest_EventTypeFilter {
				return f.GetEventTypes()
			})...))
		case *timelinepb.GetTimelineRequest_Filter_Time:
			outFilters = append(outFilters, islices.Map(filters, func(f *timelinepb.GetTimelineRequest_Filter) TimelineFilter {
				return FilterTimeRange(f.GetTime())
			})...)
		case *timelinepb.GetTimelineRequest_Filter_Id:
			outFilters = append(outFilters, islices.Map(filters, func(f *timelinepb.GetTimelineRequest_Filter) TimelineFilter {
				return FilterIDRange(f.GetId())
			})...)
		case *timelinepb.GetTimelineRequest_Filter_Call:
			outFilters = append(outFilters, FilterCall(islices.Map(filters, func(f *timelinepb.GetTimelineRequest_Filter) *timelinepb.GetTimelineRequest_CallFilter {
				return f.GetCall()
			})))
		case *timelinepb.GetTimelineRequest_Filter_Module:
			outFilters = append(outFilters, FilterModule(islices.Map(filters, func(f *timelinepb.GetTimelineRequest_Filter) *timelinepb.GetTimelineRequest_ModuleFilter {
				return f.GetModule()
			})))
		default:
			panic(fmt.Sprintf("unexpected filter type: %T", filters[0].Filter))
		}
	}
	return outFilters, ascending
}
