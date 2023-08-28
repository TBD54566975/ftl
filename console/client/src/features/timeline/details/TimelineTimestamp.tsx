import {Timestamp} from '@bufbuild/protobuf'
import {formatTimestampShort} from '../../../utils/date.utils'
import {lightTextColor} from '../../../utils/style.utils'

type Props = {
  timestamp: Timestamp
}

export const TimelineTimestamp: React.FC<Props> = ({timestamp}) => {
  return (
    <time
      dateTime={formatTimestampShort(timestamp)}
      className={`flex-none text-xs ${lightTextColor}`}
    >
      {formatTimestampShort(timestamp)}
    </time>
  )
}
