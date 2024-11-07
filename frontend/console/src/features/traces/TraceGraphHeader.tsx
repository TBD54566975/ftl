import { Activity03Icon } from 'hugeicons-react'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useRequestTraceEvents } from '../../api/timeline/use-request-trace-events'
import { totalDurationForRequest } from './traces.utils'

export const TraceGraphHeader = ({ requestKey, eventId }: { requestKey?: string; eventId: bigint }) => {
  const navigate = useNavigate()
  const requestEvents = useRequestTraceEvents(requestKey)
  const events = requestEvents.data?.reverse() ?? []

  const totalEventDuration = useMemo(() => totalDurationForRequest(events), [events])

  if (events.length === 0) {
    return null
  }

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
