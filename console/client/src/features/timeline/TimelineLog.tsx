import {LogEntry} from '../../protos/xyz/block/ftl/v1/console/console_pb'
import {formatTimestamp} from '../../utils/date.utils'
import {classNames} from '../../utils/react.utils'
import {
  logLevelBadge,
  logLevelText,
  panelColor,
  textColor,
} from '../../utils/style.utils'

type Props = {
  log: LogEntry
  selected?: boolean
}

export const TimelineLog: React.FC<Props> = ({log, selected}) => {
  return (
    <>
      <div
        className={`relative flex h-6 w-6 flex-none items-center justify-center ${panelColor}`}
      >
        <span
          className={classNames(
            `${logLevelBadge[log.logLevel]}`,
            'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium text-gray-600'
          )}
        >
          {logLevelText[log.logLevel]}
        </span>
      </div>
      <div
        className={classNames(
          `relative flex gap-x-4 flex-auto w-full max-w-full px-1 py-0.5`,
          selected && 'bg-indigo-600 rounded-md'
        )}
      >
        <div
          className={`flex-auto text-xs leading-5 ${
            selected ? 'text-white' : textColor
          } overflow-hidden overflow-ellipsis w-full max-w-full`}
        >
          <span>{log.message}</span>
        </div>

        <time
          dateTime={formatTimestamp(log.timeStamp)}
          className={`flex-none text-xs leading-5 ${
            selected ? 'text-gray-50' : 'text-gray-500'
          }`}
        >
          {formatTimestamp(log.timeStamp)}
        </time>
      </div>
    </>
  )
}
