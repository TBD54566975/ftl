import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { Event, LogEvent } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { textColor } from '../../../utils/style.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'

export const TimelineLogDetails = ({ log }: { event: Event; log: LogEvent }) => {
  return (
    <div className='p-4'>
      <div className={`p-2 text-sm bg-gray-100 dark:bg-slate-700 rounded ${textColor}`}>
        <p className='break-words whitespace-normal font-roboto-mono'>{log.message}</p>
      </div>

      <h2 className='pt-4 text-sm'>Attributes</h2>
      <CodeBlock code={JSON.stringify(log.attributes, null, 2)} language='json' />

      <DeploymentCard className='mt-4' deploymentKey={log.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        {log.requestKey && (
          <li>
            <AttributeBadge name='request' value={log.requestKey} />
          </li>
        )}
        {log.error && (
          <li>
            <AttributeBadge name='error' value={log.error} />
          </li>
        )}
      </ul>
    </div>
  )
}
