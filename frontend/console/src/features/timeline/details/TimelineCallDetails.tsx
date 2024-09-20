import type { Timestamp } from '@bufbuild/protobuf'
import { useContext } from 'react'
import { AttributeBadge } from '../../../components/AttributeBadge'
import { CloseButton } from '../../../components/CloseButton'
import { CodeBlock } from '../../../components/CodeBlock'
import type { CallEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../../providers/side-panel-provider'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { TraceGraph } from '../../traces/TraceGraph'
import { verbRefString } from '../../verbs/verb.utils'
import { TimelineTimestamp } from './TimelineTimestamp'

export const TimelineCallDetails = ({ timestamp, event }: { timestamp?: Timestamp; event: Event }) => {
  const { closePanel } = useContext(SidePanelContext)

  const call = event.entry.value as CallEvent
  return (
    <div className='p-4'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center space-x-2'>
          <div className=''>
            {call.destinationVerbRef && (
              <div className={'inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100'}>
                {verbRefString(call.destinationVerbRef)}
              </div>
            )}
          </div>
          <TimelineTimestamp timestamp={timestamp} />
        </div>
        <CloseButton onClick={closePanel} />
      </div>

      <div className='mt-4'>
        <TraceGraph requestKey={call.requestKey} selectedEventId={event.id} />
      </div>

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
    </div>
  )
}
