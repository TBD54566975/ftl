import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { AsyncExecuteEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../verbs/verb.utils'

export const TimelineAsyncExecuteDetails = ({ event }: { event: Event }) => {
  const cron = event.entry.value as AsyncExecuteEvent

  return (
    <>
      <div className='p-4'>
        {cron.error && (
          <>
            <h3>Error</h3>
            <CodeBlock code={cron.error} language='text' />
          </>
        )}

        <DeploymentCard deploymentKey={cron.deploymentKey} />

        <ul className='pt-4 space-y-2'>
          <li>
            <AttributeBadge name='duration' value={formatDuration(cron.duration)} />
          </li>
          {cron.verbRef && (
            <li>
              <AttributeBadge name='destination' value={refString(cron.verbRef)} />
            </li>
          )}
        </ul>
      </div>
    </>
  )
}
