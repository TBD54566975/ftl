import { Timestamp } from '@bufbuild/protobuf'
import { useEffect, useState } from 'react'
import { CodeBlock } from '../../../components/CodeBlock'
import { useClient } from '../../../hooks/use-client'
import { ConsoleService } from '../../../protos/xyz/block/ftl/v1/console/console_connect'
import { Call } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { getRequestCalls } from '../../../services/console.service'
import { formatDuration } from '../../../utils/date.utils'
import { textColor } from '../../../utils/style.utils'
import { RequestGraph } from '../../requests/RequestGraph'
import { verbRefString } from '../../verbs/verb.utils'
import { TimelineTimestamp } from './TimelineTimestamp'

interface Props {
  timestamp: Timestamp
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
      if (selectedCall.requestName === undefined) {
        return
      }
      const calls = await getRequestCalls(selectedCall.requestName)
      setRequestCalls(calls)
    }

    fetchRequestCalls()
  }, [client, selectedCall])

  return (
    <>
      <TimelineTimestamp timestamp={timestamp} />

      <div className='pt-2'>
        <RequestGraph calls={requestCalls} call={selectedCall} setSelectedCall={setSelectedCall} />
      </div>

      <div className='pt-4'>
        {call.destinationVerbRef && (
          <div
            className={`inline-block rounded-md dark:bg-gray-700/40 px-2 py-1 mr-1 text-xs font-medium 'text-gray-500 dark:text-gray-400 ring-1 ring-inset ring-black/10 dark:ring-white/10`}
          >
            {verbRefString(call.destinationVerbRef)}
          </div>
        )}
      </div>

      <div className='text-sm pt-2'>Request</div>
      <CodeBlock code={JSON.stringify(JSON.parse(selectedCall.request), null, 2)} language='json' />

      <div className='text-sm pt-2'>Response</div>
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
          <dd className={`${textColor}`}>{selectedCall.requestName}</dd>
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
