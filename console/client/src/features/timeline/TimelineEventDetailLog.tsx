import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { LogEntry } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { classNames } from '../../utils/react.utils'
import { logLevelBadge, logLevelText, syntaxTheme } from '../../utils/style.utils'

type Props = {
  log: LogEntry
}

export const TimelineEventDetailLog: React.FC<Props> = ({ log }) => {
  return (
    <>
      <span className={classNames(logLevelBadge[log.logLevel], 'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium text-gray-600')}>
        {logLevelText[log.logLevel]}
      </span>
      <div className='pt-4 text-sm text-gray-500 dark:text-gray-300'>
        {log.message}
      </div>
      <div className='text-sm pt-2'>
        <SyntaxHighlighter language='json'
          style={syntaxTheme()}
          customStyle={{ fontSize: '12px' }}
        >
          {JSON.stringify(log.attributes, null, 2)}
        </SyntaxHighlighter>
      </div>
    </>
  )
}
