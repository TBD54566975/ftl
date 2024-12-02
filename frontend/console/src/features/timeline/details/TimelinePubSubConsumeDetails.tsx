import { AttributeBadge } from '../../../components/AttributeBadge'
import { DeploymentCard } from '../../../features/deployments/DeploymentCard'
import { TraceGraph } from '../../../features/traces/TraceGraph'
import { TraceGraphHeader } from '../../../features/traces/TraceGraphHeader'
import type { Event, PubSubConsumeEvent } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { formatDuration } from '../../../utils/date.utils'

export const TimelinePubSubConsumeDetails = ({ event }: { event: Event }) => {
  const pubSubConsume = event.entry.value as PubSubConsumeEvent
  const destModule = `${(pubSubConsume.destVerbModule && `${pubSubConsume.destVerbModule}.`) || ''}${pubSubConsume.destVerbName || 'unknown'}`

  return (
    <>
      <div className='p-4'>
        <TraceGraphHeader requestKey={pubSubConsume.requestKey} eventId={event.id} />
        <TraceGraph requestKey={pubSubConsume.requestKey} selectedEventId={event.id} />

        <DeploymentCard className='mt-4' deploymentKey={pubSubConsume.deploymentKey} />

        <ul className='pt-4 space-y-2'>
          {pubSubConsume.requestKey && (
            <li>
              <AttributeBadge name='request' value={pubSubConsume.requestKey} />
            </li>
          )}

          <li>
            <AttributeBadge name='sink' value={destModule} />
          </li>
          <li>
            <AttributeBadge name='topic' value={pubSubConsume.topic} />
          </li>
          <li>
            <AttributeBadge name='duration' value={formatDuration(pubSubConsume.duration)} />
          </li>
        </ul>
      </div>
    </>
  )
}
