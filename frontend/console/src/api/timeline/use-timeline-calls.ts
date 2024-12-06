import { EventType } from '../../protos/xyz/block/ftl/timeline/v1/event_pb.ts'
import type { GetTimelineRequest_Filter } from '../../protos/xyz/block/ftl/timeline/v1/timeline_pb.ts'
import { eventTypesFilter } from './timeline-filters.ts'
import { useTimeline } from './use-timeline.ts'

export const useTimelineCalls = (isStreaming: boolean, filters: GetTimelineRequest_Filter[], enabled = true) => {
  const allFilters = [...filters, eventTypesFilter([EventType.CALL])]
  const timelineQuery = useTimeline(isStreaming, allFilters, 1000, enabled)

  const data = timelineQuery.data || []
  return {
    ...timelineQuery,
    data,
  }
}
