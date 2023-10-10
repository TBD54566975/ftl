import { Timestamp } from '@bufbuild/protobuf'
import { useContext } from 'react'
import { useNavigate } from 'react-router-dom'
import { AttributeBadge } from '../../../components/AttributeBadge'
import { Card } from '../../../components/Card'
import { CloseButton } from '../../../components/CloseButton'
import { CodeBlock } from '../../../components/CodeBlock'
import { Event, LogEvent } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../../providers/side-panel-provider'
import { textColor } from '../../../utils/style.utils'
import { LogLevelBadge } from '../../logs/LogLevelBadge'
import { logLevelBgColor } from '../../logs/log.utils'
import { TimelineTimestamp } from './TimelineTimestamp'

export const TimelineLogDetails = ({ event, log }: { event: Event; log: LogEvent }) => {
  const { closePanel } = useContext(SidePanelContext)
  const navigate = useNavigate()

  return (
    <>
      <div className={`${logLevelBgColor[log.logLevel]} h-2 w-full`}></div>
      <div className='p-4'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center space-x-2'>
            <LogLevelBadge logLevel={log.logLevel} />
            <TimelineTimestamp timestamp={event.timeStamp ?? new Timestamp()} />
          </div>
          <CloseButton onClick={closePanel} />
        </div>
        <div className={`mt-4 p-2 text-sm bg-gray-100 dark:bg-slate-700 rounded ${textColor}`}>
          <p className='break-words whitespace-normal font-roboto-mono'>{log.message}</p>
        </div>

        <h2 className='pt-4 text-sm'>Attributes</h2>
        <CodeBlock code={JSON.stringify(log.attributes, null, 2)} language='json' />

        <Card
          key={log.deploymentName}
          topBarColor='bg-green-500'
          className='mt-4'
          onClick={() => navigate(`/deployments/${log.deploymentName}`)}
        >
          {log.deploymentName}
          <p className='text-xs text-gray-400'>{log.deploymentName}</p>
        </Card>
        <ul className='pt-4 space-y-2'>
          {log.requestName && (
            <li>
              <AttributeBadge name='Request' value={log.requestName} />
            </li>
          )}
          {log.error && (
            <li>
              <AttributeBadge name='Error' value={log.error} />
            </li>
          )}
        </ul>
      </div>
    </>
  )
}
