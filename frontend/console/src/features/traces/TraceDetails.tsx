import type React from 'react'
import { useNavigate } from 'react-router-dom'
import type { TraceEvent } from '../../api/timeline/use-request-trace-events'
import { CallEvent, type Event, IngressEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { TimelineIcon } from '../timeline/TimelineIcon'

interface TraceDetailsProps {
  requestKey: string
  events: Event[]
  selectedEventId?: bigint
}

export const TraceDetails: React.FC<TraceDetailsProps> = ({ events, selectedEventId, requestKey }) => {
  const navigate = useNavigate()

  const firstTimeStamp = events[0]?.timeStamp
  const firstEvent = events[0].entry.value as TraceEvent
  const firstDuration = firstEvent?.duration
  const totalDurationMillis = (firstDuration?.nanos ?? 0) / 1000000

  const handleEventClick = (eventId: bigint) => {
    navigate(`/traces/${requestKey}?event_id=${eventId}`)
  }

  return (
    <div>
      <div className='mb-6 p-4 bg-gray-50 dark:bg-gray-700 rounded-lg shadow-sm'>
        <h2 className='font-semibold text-lg text-gray-800 dark:text-gray-100 mb-2'>
          Total Duration: <span className='font-bold text-indigo-600 dark:text-indigo-400'>{totalDurationMillis} ms</span>
        </h2>
        <p className='text-sm text-gray-600 dark:text-gray-300'>
          Start Time: <span className='text-gray-800 dark:text-gray-100'>{firstTimeStamp?.toDate().toLocaleString()}</span>
        </p>
      </div>

      <ul className='space-y-2'>
        {events.map((event, index) => {
          const traceEvent = event.entry.value as TraceEvent
          const durationInMillis = (traceEvent.duration?.nanos ?? 0) / 1000000

          let width = (durationInMillis / totalDurationMillis) * 100
          if (width < 1) width = 1

          const callTime = traceEvent.timeStamp?.toDate() ?? new Date()
          const initialTime = firstTimeStamp?.toDate() ?? new Date()
          const offsetInMillis = callTime.getTime() - initialTime.getTime()
          const leftOffsetPercentage = (offsetInMillis / totalDurationMillis) * 100

          let barColor = 'bg-pink-500'
          let action = ''
          let eventName = ''
          const icon = <TimelineIcon event={event} />

          if (traceEvent instanceof CallEvent) {
            barColor = 'bg-indigo-500'
            action = 'Call'
            eventName = `${traceEvent.destinationVerbRef?.module}.${traceEvent.destinationVerbRef?.name}`
          } else if (traceEvent instanceof IngressEvent) {
            barColor = 'bg-yellow-500'
            action = `HTTP ${traceEvent.method}`
            eventName = `${traceEvent.path}`
          }

          if (event.id === selectedEventId) {
            barColor = 'bg-pink-500'
          }

          const isSelected = event.id === selectedEventId
          const listItemClass = isSelected
            ? 'flex items-center justify-between p-2 bg-indigo-100/50 dark:bg-indigo-700 rounded cursor-pointer'
            : 'flex items-center justify-between p-2 hover:bg-indigo-500/10 rounded cursor-pointer'

          return (
            <li key={index} className={listItemClass} onClick={() => handleEventClick(event.id)}>
              <span className='flex items-center w-1/2 text-sm font-medium'>
                <span className='mr-2'>{icon}</span>
                <span className='mr-2'>{action}</span>
                {eventName}
              </span>

              <div className='relative w-2/3 h-4 flex-grow'>
                <div
                  className={`absolute h-4 ${barColor} rounded-sm`}
                  style={{
                    width: `${width}%`,
                    left: `${leftOffsetPercentage}%`,
                  }}
                />
              </div>
              <span className='text-xs font-medium ml-4 w-20 text-right'>{durationInMillis} ms</span>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
