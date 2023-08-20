import { LogEntry } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism'

type Props = {
  log: LogEntry
}

export const TimelineEventDetailLog: React.FC<Props> = ({ log }) => {
  return (
    <>
      <div className='text-sm text-gray-500 dark:text-gray-300'>
        {log.message}
      </div>
      <div className='text-sm pt-4'>
        <SyntaxHighlighter language='json'
          style={atomDark}
          customStyle={{ fontSize: '12px' }}
        >
          {JSON.stringify(log.attributes, null, 2)}
        </SyntaxHighlighter>
      </div>
    </>
  )
}
