import { LogEntry } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatTimestamp } from '../../utils/date.utils'
import { classNames } from '../../utils/react.utils'
import { logLevelBadge, logLevelText, panelColor, textColor } from '../../utils/style.utils'

type Props = {
  log: LogEntry
}

export const TimelineLog: React.FC<Props> = ({ log }) => {
  return (
    <>
      <div className={`relative flex h-6 w-6 flex-none items-center justify-center ${panelColor}`}>
        <span className={classNames(logLevelBadge[log.logLevel], 'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium text-gray-600')}>
          {logLevelText[log.logLevel]}
        </span>
      </div>

      <div className={`flex-auto py-0.5 text-xs leading-5 ${textColor} overflow-hidden overflow-ellipsis w-full max-w-full`}>
        <span>
          {log.message}
        </span>
      </div>

      <time
        dateTime={formatTimestamp(log.timeStamp)}
        className='flex-none py-0.5 text-xs leading-5 text-gray-500'
      >
        {formatTimestamp(log.timeStamp)}
      </time>
    </>
  )
}
