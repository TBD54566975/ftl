import { EventType, type GetEventsRequest_Filter } from '../../protos/xyz/block/ftl/console/v1/console_pb.ts'
import { eventTypesFilter } from './timeline-filters.ts'
import { useTimeline } from './use-timeline.ts'

export const useTimelineCalls = (isStreaming: boolean, filters: GetEventsRequest_Filter[], enabled = true) => {
  const allFilters = [...filters, eventTypesFilter([EventType.CALL])]
  const timelineQuery = useTimeline(isStreaming, allFilters, 1000, enabled)

  const data = timelineQuery.data || []
  return {
    ...timelineQuery,
    data,
  }
}
