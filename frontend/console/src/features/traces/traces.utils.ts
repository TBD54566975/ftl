import type { Timestamp } from '@bufbuild/protobuf'
import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'

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
