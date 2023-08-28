import {Timestamp} from '@bufbuild/protobuf'
import {CodeBlock} from '../../../components/CodeBlock'
import {
  LogEntry,
  StreamTimelineResponse,
} from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import {classNames} from '../../../utils/react.utils'
import {
  logLevelBadge,
  logLevelText,
  textColor,
} from '../../../utils/style.utils'
import {TimelineTimestamp} from './TimelineTimestamp'

type Props = {
  entry: StreamTimelineResponse
  log: LogEntry
}

export const TimelineLogDetails: React.FC<Props> = ({entry, log}) => {
  return (
    <>
      <div>
        <TimelineTimestamp timestamp={entry.timeStamp ?? new Timestamp()} />
      </div>
      <div className={`pt-4 text-xs ${textColor}`}>
        <p className='flex-wrap font-mono'>{log.message}</p>
      </div>

      <h2 className='pt-4 text-sm'>Attributes</h2>
      <CodeBlock
        code={JSON.stringify(log.attributes, null, 2)}
        language='json'
      />

      <div className='pt-2 text-gray-500 dark:text-gray-400'>
        <div className='flex pt-2 justify-between'>
          <dt>Level</dt>
          <dd className={`${textColor}`}>
            <span
              className={classNames(
                `${logLevelBadge[log.logLevel]}`,
                'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium text-gray-600'
              )}
            >
              {logLevelText[log.logLevel]}
            </span>
          </dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Deployment</dt>
          <dd className={`${textColor}`}>{log.deploymentName}</dd>
        </div>
        {log.requestKey && (
          <div className='flex pt-2 justify-between'>
            <dt>Request</dt>
            <dd className={`${textColor}`}>{log.requestKey}</dd>
          </div>
        )}
        {log.error && (
          <div className='flex pt-2 justify-between'>
            <dt>Error</dt>
            <dd className={`${textColor}`}>{log.error}</dd>
          </div>
        )}
      </div>
    </>
  )
}
