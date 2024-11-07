import { Activity03Icon } from 'hugeicons-react'
import { useNavigate } from 'react-router-dom'
import { type TraceEvent, useRequestTraceEvents } from '../../api/timeline/use-request-trace-events'
import { durationToMillis } from '../../utils'

export const TraceGraphHeader = ({ requestKey, eventId }: { requestKey?: string; eventId: bigint }) => {
  const navigate = useNavigate()
  const requestEvents = useRequestTraceEvents(requestKey)
  const events = requestEvents.data?.reverse() ?? []

  if (events.length === 0) {
    return null
  }

  const traceEvents = events.map((event) => event.entry.value as TraceEvent)
  const requestStartTime = Math.min(...traceEvents.map((event) => event.timeStamp?.toDate().getTime() ?? 0))
  const requestEndTime = Math.max(
    ...traceEvents.map((event) => {
      const eventDuration = event.duration ? durationToMillis(event.duration) : 0
      return (event.timeStamp?.toDate().getTime() ?? 0) + eventDuration
    }),
  )
  const totalEventDuration = requestEndTime - requestStartTime

  return (
    <div className='flex items-center justify-between'>
      <span className='text-xs font-mono'>
        Total <span>{totalEventDuration}ms</span>
      </span>

      <button
        type='button'
        title='View trace'
        onClick={() => navigate(`/traces/${requestKey}?event_id=${eventId}`)}
        className='flex items-center p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600 cursor-pointer'
      >
        <Activity03Icon className='w-5 h-5' />
      </button>
    </div>
  )
}
