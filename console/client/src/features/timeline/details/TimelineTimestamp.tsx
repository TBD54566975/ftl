import { Timestamp } from '@bufbuild/protobuf'
import { formatTimestampShort } from '../../../utils/date.utils'
import { lightTextColor } from '../../../utils/style.utils'

interface Props {
  timestamp?: Timestamp
}

export const TimelineTimestamp = ({ timestamp }: Props) => {
  return (
    <time dateTime={formatTimestampShort(timestamp)} className={`flex-none text-xs ${lightTextColor}`}>
      {formatTimestampShort(timestamp)}
    </time>
  )
}
