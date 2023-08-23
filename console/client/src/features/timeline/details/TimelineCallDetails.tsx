import { useEffect, useState } from 'react'
import { CodeBlock } from '../../../components/CodeBlock'
import { useClient } from '../../../hooks/use-client'
import { ConsoleService } from '../../../protos/xyz/block/ftl/v1/console/console_connect'
import { Call, StreamTimelineResponse } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatDuration } from '../../../utils/date.utils'
import { textColor } from '../../../utils/style.utils'
import { RequestGraph } from '../../requests/RequestGraph'
import { TimelineTimestamp } from './TimelineTimestamp'

type Props = {
  entry: StreamTimelineResponse
  call: Call
}

export const TimelineCallDetails: React.FC<Props> = ({ entry, call }) => {
  const client = useClient(ConsoleService)
  const [ requestCalls, setRequestCalls ] = useState<Call[]>([])

  useEffect(() => {
    const fetchRequestCalls = async () => {
      if (call.requestKey === undefined) {
        return
      }
      const response = await client.getRequestCalls({ requestKey: call.requestKey })
      setRequestCalls(response.calls)
    }
    fetchRequestCalls()
  }, [ client, call ])
  return (
    <>
      <TimelineTimestamp entry={entry} />

      <div className='pt-2'>
        <RequestGraph calls={requestCalls} call={call} />
      </div>

      <h3 className='pt-4'>Request</h3>
      <CodeBlock code={JSON.stringify(JSON.parse(call.request), null, 2)} language='json' />

      <h3 className='pt-4'>Response</h3>
      <CodeBlock code={JSON.stringify(JSON.parse(call.response), null, 2)} language='json' />

      {call.error && (
        <>
          <h3 className='pt-4'>Error</h3>
          <CodeBlock code={call.error} language='json' />
        </>
      )}

      <div className='pt-2 text-gray-500 dark:text-gray-400'>
        <div className='flex pt-2 justify-between'>
          <dt>Deployment</dt>
          <dd className={`${textColor}`}>{call.deploymentName}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Request</dt>
          <dd className={`${textColor}`}>{call.requestKey}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Duration</dt>
          <dd className={`${textColor}`}>{formatDuration(call.duration)}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Module</dt>
          <dd className={`${textColor}`}>{call.destinationVerbRef?.module}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Verb</dt>
          <dd className={`${textColor}`}>{call.destinationVerbRef?.name}</dd>
        </div>
        {call.sourceVerbRef?.module && (
          <>
            <div className='flex pt-2 justify-between'>
              <dt>Source module</dt>
              <dd className={`${textColor}`}>{call.sourceVerbRef?.module}</dd>
            </div>
            <div className='flex pt-2 justify-between'>
              <dt>Source verb</dt>
              <dd className={`${textColor}`}>{call.sourceVerbRef?.name}</dd>
            </div>
          </>
        )}
      </div>
    </>
  )
}
