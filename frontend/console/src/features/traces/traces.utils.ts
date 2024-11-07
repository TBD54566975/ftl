import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { compareTimestamps } from '../../utils'

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
