import type { Timestamp } from '@bufbuild/protobuf'
import { useContext, useEffect, useState } from 'react'
import { AttributeBadge } from '../../../components/AttributeBadge'
import { CloseButton } from '../../../components/CloseButton'
import { CodeBlock } from '../../../components/CodeBlock'
import type { IngressEvent } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../../providers/side-panel-provider'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { verbRefString } from '../../verbs/verb.utils'
import { TimelineTimestamp } from './TimelineTimestamp'

export const TimelineIngressDetails = ({ timestamp, ingress }: { timestamp?: Timestamp; ingress: IngressEvent }) => {
  const { closePanel } = useContext(SidePanelContext)
  const [selectedIngress, setSelectedIngress] = useState(ingress)

  useEffect(() => {
    setSelectedIngress(ingress)
  }, [ingress])

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

      <div className='text-sm pt-2'>Request</div>
      <CodeBlock code={JSON.stringify(JSON.parse(selectedIngress.request), null, 2)} language='json' />

      {selectedIngress.response !== 'null' && (
        <>
          <div className='text-sm pt-2'>Response</div>
          <CodeBlock code={JSON.stringify(JSON.parse(selectedIngress.response), null, 2)} language='json' />
        </>
      )}

      {selectedIngress.requestHeader !== 'null' && (
        <>
          <div className='text-sm pt-2'>Request Header</div>
          <CodeBlock code={JSON.stringify(JSON.parse(selectedIngress.requestHeader), null, 2)} language='json' />
        </>
      )}

      {selectedIngress.responseHeader !== 'null' && (
        <>
          <div className='text-sm pt-2'>Response Header</div>
          <CodeBlock code={JSON.stringify(JSON.parse(selectedIngress.responseHeader), null, 2)} language='json' />
        </>
      )}

      {selectedIngress.error && (
        <>
          <h3 className='pt-4'>Error</h3>
          <CodeBlock code={selectedIngress.error} language='text' />
        </>
      )}

      <DeploymentCard className='mt-4' deploymentKey={ingress.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='Status' value={selectedIngress.statusCode.toString()} />
        </li>
        <li>
          <AttributeBadge name='Method' value={selectedIngress.method} />
        </li>
        <li>
          <AttributeBadge name='Path' value={selectedIngress.path} />
        </li>
        {selectedIngress.requestKey && (
          <li>
            <AttributeBadge name='Request' value={selectedIngress.requestKey} />
          </li>
        )}
        <li>
          <AttributeBadge name='Duration' value={formatDuration(selectedIngress.duration)} />
        </li>
        {selectedIngress.verbRef && (
          <li>
            <AttributeBadge name='Verb' value={verbRefString(selectedIngress.verbRef)} />
          </li>
        )}
      </ul>
    </div>
  )
}
