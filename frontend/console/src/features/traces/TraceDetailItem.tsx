import type { TraceEvent } from '../../api/timeline/use-request-trace-events'
import { AsyncExecuteEvent, CallEvent, type Event, IngressEvent, PubSubPublishEvent } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { classNames } from '../../utils'
import { TimelineIcon } from '../timeline/TimelineIcon'
import { eventBackgroundColor } from '../timeline/timeline.utils'
import { eventBarLeftOffsetPercentage } from './traces.utils'

interface TraceDetailItemProps {
  event: Event
  traceEvent: TraceEvent
  eventDurationMs: number
  requestDurationMs: number
  requestStartTime: number
  selectedEventId: bigint | undefined
  handleEventClick: (eventId: bigint) => void
}

export const TraceDetailItem: React.FC<TraceDetailItemProps> = ({
  event,
  traceEvent,
  eventDurationMs,
  requestDurationMs,
  requestStartTime,
  selectedEventId,
  handleEventClick,
}) => {
  const leftOffsetPercentage = eventBarLeftOffsetPercentage(event, requestStartTime, requestDurationMs)

  let width = (eventDurationMs / requestDurationMs) * 100
  if (width < 1) width = 1

  let action = ''
  let eventName = ''
  const icon = <TimelineIcon event={event} />

  if (traceEvent instanceof CallEvent) {
    action = 'Call'
    eventName = `${traceEvent.destinationVerbRef?.module}.${traceEvent.destinationVerbRef?.name}`
  } else if (traceEvent instanceof IngressEvent) {
    action = `HTTP ${traceEvent.method}`
    eventName = `${traceEvent.path}`
  } else if (traceEvent instanceof AsyncExecuteEvent) {
    action = 'Async'
    eventName = `${traceEvent.verbRef?.module}.${traceEvent.verbRef?.name}`
  } else if (traceEvent instanceof PubSubPublishEvent) {
    action = 'Publish'
    eventName = `${traceEvent.topic}`
  }

  const barColor = event.id === selectedEventId ? 'bg-green-500' : eventBackgroundColor(event)

  const isSelected = event.id === selectedEventId
  const listItemClass = classNames(
    'flex items-center justify-between p-2 rounded cursor-pointer',
    isSelected ? 'bg-indigo-100/50 dark:bg-indigo-700' : 'hover:bg-indigo-500/10',
  )

  return (
    <li key={event.id.toString()} className={listItemClass} onClick={() => handleEventClick(event.id)}>
      <span className='flex items-center w-1/2 text-sm gap-x-2 font-medium'>
        <span>{icon}</span>
        <span>{action}</span>
        <span>{eventName}</span>
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
      <span className='text-xs font-medium ml-4 w-20 text-right'>{eventDurationMs} ms</span>
    </li>
  )
}
