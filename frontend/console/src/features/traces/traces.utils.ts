import type { TraceEvent } from '../../api/timeline/use-request-trace-events'
import type { Event } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { compareTimestamps, durationToMillis } from '../../utils'

export function eventBarLeftOffsetPercentage(event: Event, requestStartTime: number, requestDuration: number): number {
  const traceEvent = event.entry.value as TraceEvent
  // Convert bigint timestamp to number for calculation with higher precision
  const eventStartTime = (Number(traceEvent.timestamp?.seconds) * 1000000000 + (traceEvent.timestamp?.nanos ?? 0)) / 1000000 // Convert to milliseconds with nanosecond precision

  const offset = Math.max(0, eventStartTime - requestStartTime) // Ensure offset is never negative
  return Math.min(100, (offset / requestDuration) * 100) // Ensure percentage never exceeds 100
}

export const groupEventsByRequestKey = (events: Event[]): Record<string, Event[]> => {
  return events.reduce((acc: Record<string, Event[]>, event: Event) => {
    const requestKey = (() => {
      switch (event.entry.case) {
        case 'call':
        case 'ingress':
        case 'pubsubConsume':
        case 'pubsubPublish':
        case 'asyncExecute':
          return event.entry.value.requestKey
        default:
          return undefined
      }
    })()

    if (!requestKey) {
      return acc
    }

    acc[requestKey] = acc[requestKey] ? [...acc[requestKey], event] : [event]

    // Sort events by timestamp, ensuring the first event is the "trigger" event
    acc[requestKey].sort((a, b) => compareTimestamps(a.timestamp, b.timestamp))

    return acc
  }, {})
}

export const requestStartTime = (events: Event[]): number => {
  const traceEvents = events.map((event) => event.entry.value as TraceEvent)
  return Math.min(...traceEvents.map((event) => event.timestamp?.toDate().getTime() ?? 0))
}

export const totalDurationForRequest = (events: Event[]): number => {
  const traceEvents = events.map((event) => event.entry.value as TraceEvent)
  const requestEndTime = Math.max(
    ...traceEvents.map((event) => {
      const eventDuration = event.duration ? durationToMillis(event.duration) : 0
      return (event.timestamp?.toDate().getTime() ?? 0) + eventDuration
    }),
  )
  return requestEndTime - requestStartTime(events)
}
