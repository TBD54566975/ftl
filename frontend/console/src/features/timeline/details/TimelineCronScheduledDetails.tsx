import { useContext } from 'react'
import { AttributeBadge } from '../../../components/AttributeBadge'
import { CloseButton } from '../../../components/CloseButton'
import { CodeBlock } from '../../../components/CodeBlock'
import type { CronScheduledEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../../providers/side-panel-provider'
import { formatDuration, formatTimestampShort } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { verbRefString } from '../../verbs/verb.utils'
import { TimelineDetailsColorBar } from './TimelineDetailsColorBar'
import { TimelineTimestamp } from './TimelineTimestamp'

export const TimelineCronScheduledDetails = ({ event }: { event: Event }) => {
  const { closePanel } = useContext(SidePanelContext)

  const cron = event.entry.value as CronScheduledEvent

  return (
    <>
      <TimelineDetailsColorBar event={event} />
      <div className='p-4'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center space-x-2'>
            <div className=''>
              {cron.verbRef && (
                <div className={'inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100'}>
                  {verbRefString(cron.verbRef)}
                </div>
              )}
            </div>
            <TimelineTimestamp timestamp={event.timeStamp} />
          </div>
          <CloseButton onClick={closePanel} />
        </div>

        {cron.error && (
          <>
            <h3 className='pt-4'>Error</h3>
            <CodeBlock code={cron.error} language='text' />
          </>
        )}

        <DeploymentCard className='mt-4' deploymentKey={cron.deploymentKey} />

        <ul className='pt-4 space-y-2'>
          <li>
            <AttributeBadge name='Duration' value={formatDuration(cron.duration)} />
          </li>
          {cron.verbRef && (
            <li>
              <AttributeBadge name='Destination' value={verbRefString(cron.verbRef)} />
            </li>
          )}
          {cron.schedule && (
            <li>
              <AttributeBadge name='Schedule' value={cron.schedule} />
            </li>
          )}
          {cron.scheduledAt && (
            <li>
              <AttributeBadge name='Scheduled for' value={formatTimestampShort(cron.scheduledAt)} />
            </li>
          )}
        </ul>
      </div>
    </>
  )
}
