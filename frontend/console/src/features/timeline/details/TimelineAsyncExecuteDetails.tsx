import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { AsyncExecuteEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../verbs/verb.utils'
import { asyncEventTypeString } from '../timeline.utils'

export const TimelineAsyncExecuteDetails = ({ event }: { event: Event }) => {
  const asyncEvent = event.entry.value as AsyncExecuteEvent

  return (
    <>
      <div className='p-4'>
        {asyncEvent.error && (
          <>
            <h3>Error</h3>
            <CodeBlock code={asyncEvent.error} language='text' />
          </>
        )}

        <DeploymentCard deploymentKey={asyncEvent.deploymentKey} />

        <ul className='pt-4 space-y-2'>
          <li>
            <AttributeBadge name='duration' value={formatDuration(asyncEvent.duration)} />
          </li>
          <li>
            <AttributeBadge name='type' value={asyncEventTypeString(asyncEvent.asyncEventType)} />
          </li>
          {asyncEvent.verbRef && (
            <li>
              <AttributeBadge name='destination' value={refString(asyncEvent.verbRef)} />
            </li>
          )}
        </ul>
      </div>
    </>
  )
}
