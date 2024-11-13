import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { CronScheduledEvent, Event } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { formatDuration, formatTimestampShort } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../verbs/verb.utils'

export const TimelineCronScheduledDetails = ({ event }: { event: Event }) => {
  const cron = event.entry.value as CronScheduledEvent

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
          {cron.schedule && (
            <li>
              <AttributeBadge name='schedule' value={cron.schedule} />
            </li>
          )}
          {cron.scheduledAt && (
            <li>
              <AttributeBadge name='scheduled for' value={formatTimestampShort(cron.scheduledAt)} />
            </li>
          )}
        </ul>
      </div>
    </>
  )
}
