import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { Event, PubSubPublishEvent } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../verbs/verb.utils'

export const TraceDetailsPubsubPublish = ({ event }: { event: Event }) => {
  const pubsubPublish = event.entry.value as PubSubPublishEvent
  return (
    <>
      <span className='text-xl font-semibold'>Pubsub Publish Details</span>

      {pubsubPublish.error && (
        <>
          <h3 className='pt-4'>Error</h3>
          <CodeBlock code={pubsubPublish.error} language='text' />
        </>
      )}

      <DeploymentCard className='mt-4' deploymentKey={pubsubPublish.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        {pubsubPublish.requestKey && (
          <li>
            <AttributeBadge name='request' value={pubsubPublish.requestKey} />
          </li>
        )}
        <li>
          <AttributeBadge name='duration' value={formatDuration(pubsubPublish.duration)} />
        </li>
        {pubsubPublish.verbRef && (
          <li>
            <AttributeBadge name='destination' value={refString(pubsubPublish.verbRef)} />
          </li>
        )}
        {pubsubPublish.verbRef && (
          <li>
            <AttributeBadge name='source' value={refString(pubsubPublish.verbRef)} />
          </li>
        )}
      </ul>
    </>
  )
}
