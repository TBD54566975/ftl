import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { Event, PubSubPublishEvent } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../modules/decls/verb/verb.utils'

export const TraceDetailsPubsubPublish = ({ event }: { event: Event }) => {
  const pubsubPublish = event.entry.value as PubSubPublishEvent
  return (
    <>
      <span className='text-xl font-semibold'>PubSub Publish Details</span>

      {pubsubPublish.request && (
        <>
          <h3 className='pt-4'>Request</h3>
          <CodeBlock code={pubsubPublish.request} language='json' />
        </>
      )}
      {pubsubPublish.error && (
        <>
          <h3 className='pt-4'>Error</h3>
          <CodeBlock code={pubsubPublish.error} language='text' />
        </>
      )}

      <DeploymentCard className='mt-4' deploymentKey={pubsubPublish.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='topic' value={pubsubPublish.topic} />
        </li>
        <li>
          <AttributeBadge name='duration' value={formatDuration(pubsubPublish.duration)} />
        </li>
        {pubsubPublish.requestKey && (
          <li>
            <AttributeBadge name='request' value={pubsubPublish.requestKey} />
          </li>
        )}
        {pubsubPublish.verbRef && (
          <li>
            <AttributeBadge name='verb_ref' value={refString(pubsubPublish.verbRef)} />
          </li>
        )}
      </ul>
    </>
  )
}
