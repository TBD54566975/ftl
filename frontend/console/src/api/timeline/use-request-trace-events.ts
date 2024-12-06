import {
  type AsyncExecuteEvent,
  type CallEvent,
  EventType,
  type IngressEvent,
  type PubSubConsumeEvent,
  type PubSubPublishEvent,
} from '../../protos/xyz/block/ftl/timeline/v1/event_pb.ts'
import type { GetTimelineRequest_Filter } from '../../protos/xyz/block/ftl/timeline/v1/timeline_pb.ts'
import { eventTypesFilter, requestKeysFilter } from './timeline-filters.ts'
import { useTimeline } from './use-timeline.ts'

export type TraceEvent = CallEvent | IngressEvent | AsyncExecuteEvent | PubSubPublishEvent | PubSubConsumeEvent

export const useRequestTraceEvents = (requestKey?: string, filters: GetTimelineRequest_Filter[] = []) => {
  const eventTypes = [EventType.CALL, EventType.ASYNC_EXECUTE, EventType.INGRESS, EventType.PUBSUB_CONSUME, EventType.PUBSUB_PUBLISH]

  const allFilters = [...filters, requestKeysFilter([requestKey || '']), eventTypesFilter(eventTypes)]
  const timelineQuery = useTimeline(true, allFilters, 500, !!requestKey)

  const data =
    timelineQuery.data?.filter(
      (event) =>
        event.entry.case === 'call' ||
        event.entry.case === 'ingress' ||
        event.entry.case === 'asyncExecute' ||
        event.entry.case === 'pubsubPublish' ||
        event.entry.case === 'pubsubConsume',
    ) ?? []

  return {
    ...timelineQuery,
    data,
  }
}
