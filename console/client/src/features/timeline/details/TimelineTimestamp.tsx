import { StreamTimelineResponse } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatTimestampShort } from '../../../utils/date.utils'
import { lightTextColor } from '../../../utils/style.utils'

type Props = {
  entry: StreamTimelineResponse
}

export const TimelineTimestamp: React.FC<Props> = ({ entry }) => {
  return (
    <time
      dateTime={formatTimestampShort(entry.timeStamp)}
      className={`flex-none text-xs ${lightTextColor}`}
    >
      {formatTimestampShort(entry.timeStamp)}
    </time>
  )
}
