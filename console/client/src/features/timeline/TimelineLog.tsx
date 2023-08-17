import { LogEntry } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatTimestamp } from '../../utils/date.utils'
import { classNames } from '../../utils/react.utils'
import { logLevelBadge, logLevelText } from '../../utils/style.utils'

type Props = {
  log: LogEntry
}

export const TimelineLog: React.FC<Props> = ({ log }) => {
  return (
    <>
      <div className='relative flex h-6 w-6 flex-none items-center justify-center bg-white dark:bg-slate-800'>
        <span className={classNames(logLevelBadge[log.logLevel], 'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium text-gray-600')}>
          {logLevelText[log.logLevel]}
        </span>
      </div>
      <p className='flex-auto py-0.5 text-xs leading-5 text-gray-500 dark:text-gray-400'>

        <span className=' text-gray-600 dark:text-white'>
          <pre>{log.message}</pre>
        </span>

      </p>
      <time
        dateTime={formatTimestamp(log.timeStamp)}
        className='flex-none py-0.5 text-xs leading-5 text-gray-500'
      >
        {formatTimestamp(log.timeStamp)}
      </time>
    </>
  )
}
