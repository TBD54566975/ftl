import type { Duration, Timestamp } from '@bufbuild/protobuf'
import { useState } from 'react'
import { type TraceEvent, useRequestTraceEvents } from '../../api/timeline/use-request-trace-events'
import { CallEvent, IngressEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'

const EventBlock = ({
  event,
  isSelected,
  firstTimeStamp,
  firstDuration,
}: {
  event: TraceEvent
  isSelected: boolean
  firstTimeStamp: Timestamp
  firstDuration: Duration
}) => {
  const [isHovering, setIsHovering] = useState(false)

  const totalDurationMillis = (firstDuration.nanos ?? 0) / 1000000
  const durationInMillis = (event.duration?.nanos ?? 0) / 1000000
  let width = (durationInMillis / totalDurationMillis) * 100
  if (width < 1) {
    width = 1
  }

  const callTime = event.timeStamp?.toDate() ?? new Date()
  const initialTime = firstTimeStamp?.toDate() ?? new Date()
  const offsetInMillis = callTime.getTime() - initialTime.getTime()
  const leftOffsetPercentage = (offsetInMillis / totalDurationMillis) * 100

  let barColor = 'bg-pink-500'
  let eventTarget = ''

  if (event instanceof CallEvent) {
    barColor = 'bg-indigo-500'
    eventTarget = `${event.destinationVerbRef?.module}.${event.destinationVerbRef?.name}`
  } else if (event instanceof IngressEvent) {
    barColor = 'bg-yellow-500'
    eventTarget = event.path
  }

  if (isSelected) {
    barColor = 'bg-pink-500'
  }

  return (
    <div className='group relative my-0.5 h-2.5 flex' onMouseEnter={() => setIsHovering(true)} onMouseLeave={() => setIsHovering(false)}>
      <div className='flex-grow relative'>
        <div
          className={`absolute h-2.5 ${barColor} rounded-sm`}
          style={{
            width: `${width}%`,
            left: `${leftOffsetPercentage}%`,
          }}
        />
        {isHovering && (
          <div className='absolute top-[-40px] right-0 bg-gray-100 dark:bg-gray-700  text-xs p-2 rounded shadow-lg z-10 w-max flex flex-col items-end'>
            <p>
              {event instanceof CallEvent ? 'Call ' : 'Ingress '}
              <span className='text-indigo-500 dark:text-indigo-400'>{eventTarget}</span>
              {` (${durationInMillis} ms)`}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}

export const TraceGraph = ({ requestKey, selectedEventId }: { requestKey?: string; selectedEventId?: bigint }) => {
  const requestEvents = useRequestTraceEvents(requestKey)
  const events = requestEvents.data?.reverse() ?? []

  if (events.length === 0) {
    return
  }

  const firstTimeStamp = events[0].timeStamp
  const traceEvent = events[0].entry.value as TraceEvent
  const firstDuration = traceEvent.duration
  if (firstTimeStamp === undefined || firstDuration === undefined) {
    return
  }

  return (
    <div className='flex flex-col'>
      {events.map((c, index) => (
        <div key={index} className='flex hover:bg-indigo-500/60 hover:dark:bg-indigo-500/10 rounded-sm'>
          <div className='w-full relative'>
            <EventBlock
              event={c.entry.value as TraceEvent}
              isSelected={c.id === selectedEventId}
              firstTimeStamp={firstTimeStamp}
              firstDuration={firstDuration}
            />
          </div>
        </div>
      ))}
    </div>
  )
}
