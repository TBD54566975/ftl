import { useEffect, useState } from 'react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { useClient } from '../../hooks/use-client'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { Call } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { RequestGraph } from '../requests/RequestGraph'
import { syntaxTheme } from '../../utils/style.utils'

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
      <RequestGraph calls={requestCalls} call={call} />

      <h3 className='pt-4'>Request</h3>
      <div className='text-sm'>
        <SyntaxHighlighter language='json'
          style={syntaxTheme}
          customStyle={{ fontSize: '12px' }}
        >
          {JSON.stringify(JSON.parse(call.request), null, 2)}
        </SyntaxHighlighter>
      </div>
      <h3 className='pt-4'>Response</h3>
      <div className='text-sm'>
        <SyntaxHighlighter language='json'
          style={syntaxTheme}
          customStyle={{ fontSize: '12px' }}
        >
          {JSON.stringify(JSON.parse(call.response), null, 2)}
        </SyntaxHighlighter>
      </div>
    </>
  )
}
