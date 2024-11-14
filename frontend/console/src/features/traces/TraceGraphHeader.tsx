import { Activity03Icon } from 'hugeicons-react'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useRequestTraceEvents } from '../../api/timeline/use-request-trace-events'
import { Button } from '../../components/Button'
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
    <div className='flex items-center justify-between mb-1'>
      <span className='text-xs font-mono'>
        Total <span>{totalEventDuration}ms</span>
      </span>

      <Button variant='secondary' size='sm' onClick={() => navigate(`/traces/${requestKey}?event_id=${eventId}`)} title='View trace'>
        <Activity03Icon className='size-5' />
      </Button>
    </div>
  )
}
