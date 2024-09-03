import type { Timestamp } from '@bufbuild/protobuf'
import { formatTimestampShort } from '../../../utils/date.utils'
import { textColor } from '../../../utils/style.utils'

export const TimelineTimestamp = ({ timestamp }: { timestamp?: Timestamp }) => {
  return (
    <time dateTime={formatTimestampShort(timestamp)} className={`flex-none text-sm ${textColor}`}>
      {formatTimestampShort(timestamp)}
    </time>
  )
}
