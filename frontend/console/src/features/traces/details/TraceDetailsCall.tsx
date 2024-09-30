import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { CallEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { verbRefString } from '../../verbs/verb.utils'

export const TraceDetailsCall = ({ event }: { event: Event }) => {
  const call = event.entry.value as CallEvent
  return (
    <>
      <span className='text-xl font-semibold'>Call Details</span>
      <div className='text-sm pt-2'>Request</div>
      <CodeBlock code={JSON.stringify(JSON.parse(call.request), null, 2)} language='json' />

      {call.response !== 'null' && (
        <>
          <div className='text-sm pt-2'>Response</div>
          <CodeBlock code={JSON.stringify(JSON.parse(call.response), null, 2)} language='json' />
        </>
      )}

      {call.error && (
        <>
          <h3 className='pt-4'>Error</h3>
          <CodeBlock code={call.error} language='text' />
          {call.stack && (
            <>
              <h3 className='pt-4'>Stack</h3>
              <CodeBlock code={call.stack} language='text' />
            </>
          )}
        </>
      )}

      <DeploymentCard className='mt-4' deploymentKey={call.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        {call.requestKey && (
          <li>
            <AttributeBadge name='Request' value={call.requestKey} />
          </li>
        )}
        <li>
          <AttributeBadge name='Duration' value={formatDuration(call.duration)} />
        </li>
        {call.destinationVerbRef && (
          <li>
            <AttributeBadge name='Destination' value={verbRefString(call.destinationVerbRef)} />
          </li>
        )}
        {call.sourceVerbRef && (
          <li>
            <AttributeBadge name='Source' value={verbRefString(call.sourceVerbRef)} />
          </li>
        )}
      </ul>
    </>
  )
}
