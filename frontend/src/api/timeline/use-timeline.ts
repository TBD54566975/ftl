import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useClient } from '../../hooks/use-client'
import { useVisibility } from '../../hooks/use-visibility'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { type EventsQuery_Filter, EventsQuery_Order } from '../../protos/xyz/block/ftl/v1/console/console_pb'

const timelineKey = 'timeline'
const maxTimelineEntries = 1000

export const useTimeline = (isStreaming: boolean, filters: EventsQuery_Filter[], enabled = true) => {
  const client = useClient(ConsoleService)
  const queryClient = useQueryClient()
  const isVisible = useVisibility()

  const order = EventsQuery_Order.DESC
  const limit = isStreaming ? 200 : 1000

  const queryKey = [timelineKey, isStreaming, filters, order, limit]

  const fetchTimeline = async ({ signal }: { signal: AbortSignal }) => {
    try {
      console.log('fetching timeline')
      const response = await client.getEvents({ filters, limit, order }, { signal })
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
      console.log('streaming timeline')
      console.log('filters:', filters)
      for await (const response of client.streamEvents({ updateInterval: { seconds: BigInt(1) }, query: { limit, filters, order } }, { signal })) {
        if (response.events) {
          const prev = queryClient.getQueryData<Event[]>(queryKey) ?? []
          const allEvents = [...response.events, ...prev].slice(0, maxTimelineEntries)
          queryClient.setQueryData(queryKey, allEvents)
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
