import type React from 'react'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import type { TraceEvent } from '../../api/timeline/use-request-trace-events'
import type { Event } from '../../protos/xyz/block/ftl/v1/event_pb'
import { TraceDetailItem } from './TraceDetailItem'
import { TraceRulerItem } from './TraceRulerItem'
import { requestStartTime, totalDurationForRequest } from './traces.utils'

interface TraceDetailsProps {
  requestKey: string
  events: Event[]
  selectedEventId?: bigint
}

export const TraceDetails: React.FC<TraceDetailsProps> = ({ events, selectedEventId, requestKey }) => {
  const navigate = useNavigate()

  const startTime = useMemo(() => requestStartTime(events), [events])
  const totalEventDuration = useMemo(() => totalDurationForRequest(events), [events])

  const handleEventClick = (eventId: bigint) => {
    navigate(`/traces/${requestKey}?event_id=${eventId}`)
  }

  return (
    <div>
      <div className='mb-6 p-4 bg-gray-50 dark:bg-gray-700 rounded-lg shadow-sm'>
        <h2 className='font-semibold text-lg text-gray-800 dark:text-gray-100 mb-2'>
          Total Duration: <span className='font-bold text-indigo-600 dark:text-indigo-400'>{totalEventDuration} ms</span>
        </h2>
        <p className='text-sm text-gray-600 dark:text-gray-300'>
          Start Time: <span className='text-gray-800 dark:text-gray-100'>{new Date(startTime).toLocaleString()}</span>
        </p>
      </div>

      <ul>
        <div className='mb-1'>
          <TraceRulerItem duration={totalEventDuration} />
        </div>

        {events.map((event, index) => {
          const traceEvent = event.entry.value as TraceEvent
          const eventDurationMs = (traceEvent.duration?.nanos ?? 0) / 1000000

          return (
            <TraceDetailItem
              key={index}
              event={event}
              traceEvent={traceEvent}
              eventDurationMs={eventDurationMs}
              requestDurationMs={totalEventDuration}
              requestStartTime={startTime}
              selectedEventId={selectedEventId}
              handleEventClick={handleEventClick}
            />
          )
        })}
      </ul>
    </div>
  )
}
