import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { Event, IngressEvent } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../verbs/verb.utils'

export const TraceDetailsIngress = ({ event }: { event: Event }) => {
  const ingress = event.entry.value as IngressEvent
  return (
    <>
      <span className='text-xl font-semibold'>Call Details</span>
      <div className='text-sm pt-2'>Request</div>
      <CodeBlock code={JSON.stringify(JSON.parse(ingress.request), null, 2)} language='json' />

      {ingress.response !== 'null' && (
        <>
          <div className='text-sm pt-2'>Response</div>
          <CodeBlock code={JSON.stringify(JSON.parse(ingress.response), null, 2)} language='json' />
        </>
      )}

      {ingress.requestHeader !== 'null' && (
        <>
          <div className='text-sm pt-2'>Request Header</div>
          <CodeBlock code={JSON.stringify(JSON.parse(ingress.requestHeader), null, 2)} language='json' />
        </>
      )}

      {ingress.responseHeader !== 'null' && (
        <>
          <div className='text-sm pt-2'>Response Header</div>
          <CodeBlock code={JSON.stringify(JSON.parse(ingress.responseHeader), null, 2)} language='json' />
        </>
      )}

      {ingress.error && (
        <>
          <h3 className='pt-4'>Error</h3>
          <CodeBlock code={ingress.error} language='text' />
        </>
      )}

      <DeploymentCard className='mt-4' deploymentKey={ingress.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='status' value={ingress.statusCode.toString()} />
        </li>
        <li>
          <AttributeBadge name='method' value={ingress.method} />
        </li>
        <li>
          <AttributeBadge name='path' value={ingress.path} />
        </li>
        {ingress.requestKey && (
          <li>
            <AttributeBadge name='request' value={ingress.requestKey} />
          </li>
        )}
        <li>
          <AttributeBadge name='duration' value={formatDuration(ingress.duration)} />
        </li>
        {ingress.verbRef && (
          <li>
            <AttributeBadge name='verb' value={refString(ingress.verbRef)} />
          </li>
        )}
      </ul>
    </>
  )
}
