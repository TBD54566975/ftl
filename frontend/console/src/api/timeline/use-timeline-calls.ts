import { type CallEvent, EventType, type EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import { eventTypesFilter } from './timeline-filters.ts'
import { useTimeline } from './use-timeline.ts'

export const useTimelineCalls = (isStreaming: boolean, filters: EventsQuery_Filter[], enabled = true) => {
  const allFilters = [...filters, eventTypesFilter([EventType.CALL])]
  const timelineQuery = useTimeline(isStreaming, allFilters, enabled)

  // Map the events to CallEvent for ease of use
  const data = timelineQuery.data?.map((event) => event.entry.value as CallEvent) || []

  return {
    ...timelineQuery,
    data,
  }
}
