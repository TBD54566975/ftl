import { Call } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { useClient } from '../../hooks/use-client'
import { useEffect, useState } from 'react'
import { RequestGraph } from '../requests/RequestGraph'

type Props = {
  call: Call
}

export const TimelineEventDetailCall: React.FC<Props> = ({ call }) => {
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
      <RequestGraph calls={requestCalls} />
    </>
  )
}
