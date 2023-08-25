import { Timestamp } from '@bufbuild/protobuf'
import { useEffect, useState } from 'react'
import { CodeBlock } from '../../../components/CodeBlock'
import { useClient } from '../../../hooks/use-client'
import { ConsoleService } from '../../../protos/xyz/block/ftl/v1/console/console_connect'
import { Call } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatDuration } from '../../../utils/date.utils'
import { textColor } from '../../../utils/style.utils'
import { RequestGraph } from '../../requests/RequestGraph'
import { TimelineTimestamp } from './TimelineTimestamp'

interface Props {
  timestamp?: Timestamp
  call: Call
}

export const TimelineCallDetails = ({ timestamp, call }: Props) => {
  const client = useClient(ConsoleService)
  const [requestCalls, setRequestCalls] = useState<Call[]>([])
  const [selectedCall, setSelectedCall] = useState(call)

  useEffect(() => {
    setSelectedCall(call)
  }, [call])

  useEffect(() => {
    const fetchRequestCalls = async () => {
      if (selectedCall.requestKey === undefined) {
        return
      }
      const response = await client.getRequestCalls({
        requestKey: selectedCall.requestKey,
      })
      setRequestCalls(response.calls)
    }
    fetchRequestCalls()
  }, [client, selectedCall])

  return (
    <>
      <TimelineTimestamp timestamp={timestamp} />

      <div className='pt-2'>
        <RequestGraph calls={requestCalls} call={selectedCall} setSelectedCall={setSelectedCall} />
      </div>

      <h3 className='pt-4'>Request</h3>
      <CodeBlock code={JSON.stringify(JSON.parse(selectedCall.request), null, 2)} language='json' />

      <h3 className='pt-4'>Response</h3>
      <CodeBlock code={JSON.stringify(JSON.parse(selectedCall.response), null, 2)} language='json' />

      {selectedCall.error && (
        <>
          <h3 className='pt-4'>Error</h3>
          <CodeBlock code={selectedCall.error} language='json' />
        </>
      )}

      <div className='pt-2 text-gray-500 dark:text-gray-400'>
        <div className='flex pt-2 justify-between'>
          <dt>Deployment</dt>
          <dd className={`${textColor}`}>{selectedCall.deploymentName}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Request</dt>
          <dd className={`${textColor}`}>{selectedCall.requestKey}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Duration</dt>
          <dd className={`${textColor}`}>{formatDuration(selectedCall.duration)}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Module</dt>
          <dd className={`${textColor}`}>{selectedCall.destinationVerbRef?.module}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Verb</dt>
          <dd className={`${textColor}`}>{selectedCall.destinationVerbRef?.name}</dd>
        </div>
        {selectedCall.sourceVerbRef?.module && (
          <>
            <div className='flex pt-2 justify-between'>
              <dt>Source module</dt>
              <dd className={`${textColor}`}>{selectedCall.sourceVerbRef?.module}</dd>
            </div>
            <div className='flex pt-2 justify-between'>
              <dt>Source verb</dt>
              <dd className={`${textColor}`}>{selectedCall.sourceVerbRef?.name}</dd>
            </div>
          </>
        )}
      </div>
    </>
  )
}
