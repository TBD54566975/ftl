import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { AsyncExecuteEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { TraceGraph } from '../../traces/TraceGraph'
import { TraceGraphHeader } from '../../traces/TraceGraphHeader'
import { refString } from '../../verbs/verb.utils'
import { asyncEventTypeString } from '../timeline.utils'

export const TimelineAsyncExecuteDetails = ({ event }: { event: Event }) => {
  const asyncEvent = event.entry.value as AsyncExecuteEvent

  return (
    <>
      <div className='p-4'>
        <div className='pb-2'>
          <TraceGraphHeader requestKey={asyncEvent.requestKey} eventId={event.id} />
          <TraceGraph requestKey={asyncEvent.requestKey} selectedEventId={event.id} />
        </div>

        {asyncEvent.error && (
          <>
            <h3>Error</h3>
            <CodeBlock code={asyncEvent.error} language='text' />
          </>
        )}
        <DeploymentCard deploymentKey={asyncEvent.deploymentKey} />
        <ul className='pt-4 space-y-2'>
          {asyncEvent.requestKey && (
            <li>
              <AttributeBadge name='request' value={asyncEvent.requestKey} />
            </li>
          )}
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
