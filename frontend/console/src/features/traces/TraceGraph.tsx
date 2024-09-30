import type { Duration, Timestamp } from '@bufbuild/protobuf'
import { useState } from 'react'
import { type TraceEvent, useRequestTraceEvents } from '../../api/timeline/use-request-trace-events'
import { CallEvent, type Event, IngressEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { classNames } from '../../utils'
import { eventBackgroundColor } from '../timeline/timeline.utils'
import { eventBarLeftOffsetPercentage } from './traces.utils'

const EventBlock = ({
  event,
  isSelected,
  requestStartTime,
  requestDuration,
}: {
  event: Event
  isSelected: boolean
  requestStartTime: Timestamp
  requestDuration: Duration
}) => {
  const [isHovering, setIsHovering] = useState(false)

  const traceEvent = event.entry.value as TraceEvent
  const totalDurationMillis = (requestDuration.nanos ?? 0) / 1000000
  const durationInMillis = (traceEvent.duration?.nanos ?? 0) / 1000000
  let width = (durationInMillis / totalDurationMillis) * 100
  if (width < 1) {
    width = 1
  }

  const leftOffsetPercentage = eventBarLeftOffsetPercentage(event, requestStartTime, totalDurationMillis)

  let eventTarget = ''

  if (traceEvent instanceof CallEvent) {
    eventTarget = `${traceEvent.destinationVerbRef?.module}.${traceEvent.destinationVerbRef?.name}`
  } else if (traceEvent instanceof IngressEvent) {
    eventTarget = traceEvent.path
  }

  const barColor = isSelected ? 'bg-green-500' : eventBackgroundColor(event)

  return (
    <div className='group relative my-0.5 h-2.5 flex' onMouseEnter={() => setIsHovering(true)} onMouseLeave={() => setIsHovering(false)}>
      <div className='flex-grow relative'>
        <div
          className={classNames('absolute h-2.5 rounded-sm', barColor)}
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

  const requestStartTime = events[0].timeStamp
  const traceEvent = events[0].entry.value as TraceEvent
  const firstEventDuration = traceEvent.duration
  if (requestStartTime === undefined || firstEventDuration === undefined) {
    return
  }

  return (
    <div className='flex flex-col'>
      {events.map((c, index) => (
        <div key={index} className='flex hover:bg-indigo-500/60 hover:dark:bg-indigo-500/10 rounded-sm'>
          <div className='w-full relative'>
            <EventBlock event={c} isSelected={c.id === selectedEventId} requestStartTime={requestStartTime} requestDuration={firstEventDuration} />
          </div>
        </div>
      ))}
    </div>
  )
}
