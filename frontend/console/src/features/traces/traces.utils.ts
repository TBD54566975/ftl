import type { TraceEvent } from '../../api/timeline/use-request-trace-events'
import type { Event } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { compareTimestamps, durationToMillis } from '../../utils'

export const eventBarLeftOffsetPercentage = (event: Event, requestStartTime: number, requestDurationMs: number) => {
  if (!event.timeStamp) {
    return 0
  }

  const offsetInMillis = event.timeStamp.toDate().getTime() - requestStartTime
  return (offsetInMillis / requestDurationMs) * 100
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
    acc[requestKey].sort((a, b) => compareTimestamps(a.timeStamp, b.timeStamp))

    return acc
  }, {})
}

export const requestStartTime = (events: Event[]): number => {
  const traceEvents = events.map((event) => event.entry.value as TraceEvent)
  return Math.min(...traceEvents.map((event) => event.timeStamp?.toDate().getTime() ?? 0))
}

export const totalDurationForRequest = (events: Event[]): number => {
  const traceEvents = events.map((event) => event.entry.value as TraceEvent)
  const requestEndTime = Math.max(
    ...traceEvents.map((event) => {
      const eventDuration = event.duration ? durationToMillis(event.duration) : 0
      return (event.timeStamp?.toDate().getTime() ?? 0) + eventDuration
    }),
  )
  return requestEndTime - requestStartTime(events)
}
