import type React from 'react'
import { useNavigate } from 'react-router-dom'
import type { TraceEvent } from '../../api/timeline/use-request-trace-events'
import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { TraceDetailItem } from './TraceDetailItem'

interface TraceDetailsProps {
  requestKey: string
  events: Event[]
  selectedEventId?: bigint
}

export const TraceDetails: React.FC<TraceDetailsProps> = ({ events, selectedEventId, requestKey }) => {
  const navigate = useNavigate()

  const requestStartTime = events[0]?.timeStamp
  const firstEvent = events[0].entry.value as TraceEvent
  const requestDurationMs = (firstEvent?.duration?.nanos ?? 0) / 1000000

  const handleEventClick = (eventId: bigint) => {
    navigate(`/traces/${requestKey}?event_id=${eventId}`)
  }

  return (
    <div>
      <div className='mb-6 p-4 bg-gray-50 dark:bg-gray-700 rounded-lg shadow-sm'>
        <h2 className='font-semibold text-lg text-gray-800 dark:text-gray-100 mb-2'>
          Total Duration: <span className='font-bold text-indigo-600 dark:text-indigo-400'>{requestDurationMs} ms</span>
        </h2>
        <p className='text-sm text-gray-600 dark:text-gray-300'>
          Start Time: <span className='text-gray-800 dark:text-gray-100'>{requestStartTime?.toDate().toLocaleString()}</span>
        </p>
      </div>

      <ul className='space-y-2'>
        {events.map((event, index) => {
          const traceEvent = event.entry.value as TraceEvent
          const eventDurationMs = (traceEvent.duration?.nanos ?? 0) / 1000000

          return (
            <TraceDetailItem
              key={index}
              event={event}
              traceEvent={traceEvent}
              eventDurationMs={eventDurationMs}
              requestDurationMs={requestDurationMs}
              requestStartTime={requestStartTime}
              selectedEventId={selectedEventId}
              handleEventClick={handleEventClick}
            />
          )
        })}
      </ul>
    </div>
  )
}
