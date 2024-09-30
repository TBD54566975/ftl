import { EventType, type EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import { eventTypesFilter, moduleFilter } from './timeline-filters.ts'
import { useTimeline } from './use-timeline.ts'

export const useModuleTraceEvents = (module: string, verb?: string, filters: EventsQuery_Filter[] = []) => {
  const eventTypes = [EventType.CALL, EventType.INGRESS]
  const allFilters = [...filters, moduleFilter(module, verb), eventTypesFilter(eventTypes)]
  const timelineQuery = useTimeline(true, allFilters, 500)

  const data = timelineQuery.data?.filter((event) => event.entry.case === 'call' || event.entry.case === 'ingress') ?? []
  return {
    ...timelineQuery,
    data,
  }
}
