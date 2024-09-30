import { type CallEvent, EventType, type EventsQuery_Filter, type IngressEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import { eventTypesFilter, requestKeysFilter } from './timeline-filters.ts'
import { useTimeline } from './use-timeline.ts'

export type TraceEvent = CallEvent | IngressEvent

export const useRequestTraceEvents = (requestKey?: string, filters: EventsQuery_Filter[] = []) => {
  const eventTypes = [EventType.CALL, EventType.INGRESS]
  const allFilters = [...filters, requestKeysFilter([requestKey || '']), eventTypesFilter(eventTypes)]
  const timelineQuery = useTimeline(true, allFilters, 500, !!requestKey)

  const data = timelineQuery.data?.filter((event) => event.entry.case === 'call' || event.entry.case === 'ingress') ?? []
  return {
    ...timelineQuery,
    data,
  }
}
