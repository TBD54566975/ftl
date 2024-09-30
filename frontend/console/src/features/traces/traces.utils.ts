import type { Timestamp } from '@bufbuild/protobuf'
import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { compareTimestamps } from '../../utils'

export const eventBarLeftOffsetPercentage = (event: Event, requestStartTime: Timestamp | undefined, requestDurationMs: number) => {
  if (!requestStartTime) {
    return 0
  }

  if (!event.timeStamp || !requestStartTime) {
    return 0
  }

  const callTime = event.timeStamp.toDate()
  const initialTime = requestStartTime.toDate()
  const offsetInMillis = callTime.getTime() - initialTime.getTime()

  return (offsetInMillis / requestDurationMs) * 100
}

export const groupEventsByRequestKey = (events: Event[]): Record<string, Event[]> => {
  const even = events.reduce((acc: Record<string, Event[]>, event: Event) => {
    let requestKey: string | undefined

    if (event.entry.case === 'call') {
      requestKey = event.entry.value.requestKey
    } else if (event.entry.case === 'ingress') {
      requestKey = event.entry.value.requestKey
    }

    if (requestKey) {
      if (!acc[requestKey]) {
        acc[requestKey] = []
      }

      acc[requestKey].push(event)

      // first event will be the "trigger" event
      acc[requestKey].sort((a, b) => compareTimestamps(a.timeStamp, b.timeStamp))
    }

    return acc
  }, {})

  return even
}
