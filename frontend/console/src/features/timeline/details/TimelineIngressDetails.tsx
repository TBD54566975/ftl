import type { Timestamp } from '@bufbuild/protobuf'
import { useContext } from 'react'
import { AttributeBadge } from '../../../components/AttributeBadge'
import { CloseButton } from '../../../components/CloseButton'
import { CodeBlock } from '../../../components/CodeBlock'
import type { Event, IngressEvent } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../../providers/side-panel-provider'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { TraceGraph } from '../../traces/TraceGraph'
import { verbRefString } from '../../verbs/verb.utils'
import { TimelineTimestamp } from './TimelineTimestamp'

export const TimelineIngressDetails = ({ timestamp, event }: { timestamp?: Timestamp; event: Event }) => {
  const { closePanel } = useContext(SidePanelContext)

  const ingress = event.entry.value as IngressEvent
  return (
    <div className='p-4'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center space-x-2'>
          <div className=''>
            {ingress.verbRef && (
              <div className={'inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100'}>
                {`${ingress.method} ${ingress.path}`}
              </div>
            )}
          </div>
          <TimelineTimestamp timestamp={timestamp} />
        </div>
        <CloseButton onClick={closePanel} />
      </div>

      <div className='mt-4'>
        <TraceGraph requestKey={ingress.requestKey} selectedEventId={event.id} />
      </div>

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
          <AttributeBadge name='Status' value={ingress.statusCode.toString()} />
        </li>
        <li>
          <AttributeBadge name='Method' value={ingress.method} />
        </li>
        <li>
          <AttributeBadge name='Path' value={ingress.path} />
        </li>
        {ingress.requestKey && (
          <li>
            <AttributeBadge name='Request' value={ingress.requestKey} />
          </li>
        )}
        <li>
          <AttributeBadge name='Duration' value={formatDuration(ingress.duration)} />
        </li>
        {ingress.verbRef && (
          <li>
            <AttributeBadge name='Verb' value={verbRefString(ingress.verbRef)} />
          </li>
        )}
      </ul>
    </div>
  )
}
