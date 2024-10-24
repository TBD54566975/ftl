import { Activity03Icon } from 'hugeicons-react'
import { useNavigate } from 'react-router-dom'
import { type TraceEvent, useRequestTraceEvents } from '../../api/timeline/use-request-trace-events'

export const TraceGraphHeader = ({ requestKey, eventId }: { requestKey?: string; eventId: bigint }) => {
  const navigate = useNavigate()
  const requestEvents = useRequestTraceEvents(requestKey)
  const events = requestEvents.data?.reverse() ?? []

  if (events.length === 0) {
    return null
  }

  const firstTimeStamp = events[0].timeStamp
  const traceEvent = events[0].entry.value as TraceEvent
  const firstDuration = traceEvent.duration
  if (firstTimeStamp === undefined || firstDuration === undefined) {
    return null
  }

  const totalDurationMillis = (firstDuration.nanos ?? 0) / 1000000

  return (
    <div className='flex items-center justify-between'>
      <span className='text-xs font-mono'>
        Total <span>{totalDurationMillis}ms</span>
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
