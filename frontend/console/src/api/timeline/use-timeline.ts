import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useClient } from '../../hooks/use-client'
import { useVisibility } from '../../hooks/use-visibility'
import { ConsoleService } from '../../protos/xyz/block/ftl/console/v1/console_connect'
import type { Event } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { type GetTimelineRequest_Filter, GetTimelineRequest_Order } from '../../protos/xyz/block/ftl/timeline/v1/timeline_pb'
import { compareTimestamps } from '../../utils/date.utils'

const timelineKey = 'timeline'
const maxTimelineEntries = 1000

export const useTimeline = (isStreaming: boolean, filters: GetTimelineRequest_Filter[], updateIntervalMs = 1000, enabled = true) => {
  const client = useClient(ConsoleService)
  const queryClient = useQueryClient()
  const isVisible = useVisibility()

  const order = GetTimelineRequest_Order.DESC
  const limit = isStreaming ? 200 : 1000

  const queryKey = [timelineKey, isStreaming, filters, order, limit]

  const fetchTimeline = async ({ signal }: { signal: AbortSignal }) => {
    try {
      console.debug('fetching timeline')
      const response = await client.getTimeline({ filters, limit, order }, { signal })
      return response.events
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.Canceled) {
          return []
        }
      }
      throw error
    }
  }

  const streamTimeline = async ({ signal }: { signal: AbortSignal }) => {
    try {
      console.debug('streaming timeline')
      console.debug('timeline-filters:', filters)

      // Clear the cache when starting a new stream
      queryClient.setQueryData<Event[]>(queryKey, (_ = []) => [])

      for await (const response of client.streamTimeline(
        { updateInterval: { seconds: BigInt(0), nanos: updateIntervalMs * 1000 }, query: { limit, filters, order } },
        { signal },
      )) {
        console.debug('timeline-response:', response)
        if (response.events) {
          queryClient.setQueryData<Event[]>(queryKey, (prev = []) => {
            return [...response.events, ...prev].sort((a, b) => compareTimestamps(b.timestamp, a.timestamp)).slice(0, maxTimelineEntries)
          })
        }
      }
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code !== Code.Canceled) {
          console.error('Console service - streamEvents - Connect error:', error)
        }
      } else {
        console.error('Console service - streamEvents:', error)
      }
    }
  }

  return useQuery({
    queryKey: queryKey,
    queryFn: async ({ signal }) => (isStreaming ? streamTimeline({ signal }) : fetchTimeline({ signal })),
    enabled: enabled && isVisible,
  })
}
