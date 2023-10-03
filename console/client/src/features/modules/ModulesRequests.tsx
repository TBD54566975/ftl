import React from 'react'
import { Timestamp } from '@bufbuild/protobuf'
import { Disclosure } from '@headlessui/react'
import { formatDuration, formatTimestamp } from '../../utils'
import {
  AttributeBadge,
  CodeBlock
} from '../../components'
import { Panel } from './components'
import { TimelineTimestamp } from '../timeline/details/TimelineTimestamp'
import { VerbId } from './modules.constants'
import { useClient } from '../../hooks/use-client'
import { CallEvent, Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { getCalls, getRequestCalls } from '../../services/console.service'
import { RequestGraph } from '../requests/RequestGraph'
import { verbRefString, getNames } from './modules.utils'

const RequestDetails: React.FC<{ timestamp: Timestamp, call: CallEvent}> = ({ call, timestamp }) => {
  const client = useClient(ConsoleService)
  const [requestCalls, setRequestCalls] = React.useState<CallEvent[]>([])
  const [selectedCall, setSelectedCall] = React.useState(call)

  React.useEffect(() => {
    setSelectedCall(call)
  }, [call])

  React.useEffect(() => {
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
    <div className='p-4'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center space-x-2'>
          <div className=''>
            {call.destinationVerbRef && (
              <div
                className={`inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100`}
              >
                {verbRefString(call.destinationVerbRef)}
              </div>
            )}
          </div>
          <TimelineTimestamp timestamp={timestamp} />
        </div>
      </div>

      <div className='pt-4'>
        <RequestGraph calls={requestCalls} call={selectedCall} setSelectedCall={setSelectedCall} />
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

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='Deployment' value={selectedCall.deploymentName} />
        </li>
        {selectedCall.requestName && (
          <li>
            <AttributeBadge name='Request' value={selectedCall.requestName} />
          </li>
        )}
        <li>
          <AttributeBadge name='Duration' value={formatDuration(selectedCall.duration)} />
        </li>
        {selectedCall.destinationVerbRef && (
          <li>
            <AttributeBadge name='Destination' value={verbRefString(selectedCall.destinationVerbRef)} />
          </li>
        )}
        {selectedCall.sourceVerbRef && (
          <li>
            <AttributeBadge name='Source' value={verbRefString(selectedCall.sourceVerbRef)} />
          </li>
        )}
      </ul>
    </div>
  )
}

const RequestRow:React.FC<{ call: CallEvent }> = ({call}) => {
  return(
    <div className=''>
      <Disclosure>
        {({ open }) => (
          <>
            <Disclosure.Button>
              <div className='truncate text-sm font-medium leading-6 text-gray-700 dark:text-white'>
                {call.sourceVerbRef && [call.sourceVerbRef.module, call.sourceVerbRef.name].join(':')}
              </div>
              <div className='font-mono text-sm leading-6 text-gray-500 dark:text-gray-400'>
                {formatTimestamp(call.timeStamp)}
              </div>
              <div
                className={`text-right text-sm leading-6 text-gray-500 dark:text-gray-400`}
              >
                {formatDuration(call.duration)}
              </div>
            </Disclosure.Button>
            <Disclosure.Panel>
              <RequestDetails call={call} timestamp={call.timeStamp ?? new Timestamp()} />
            </Disclosure.Panel>
          </>
        )}
      </Disclosure>
    </div>
  )
}

export const ModulesRequests: React.FC<{
  className: string
  modules: Module[]
  selectedVerbs: VerbId[]
}> = ({
  className,
  modules,
  selectedVerbs,
}) => {
  if (!selectedVerbs.length) return <></>
  const client = useClient(ConsoleService)
  const [calls, setCalls] = React.useState<CallEvent[]>([])
  React.useEffect(() => {
    const fetchCalls = async () => {
      const verbs: [string, Verb][] = []
      for(const verbId of selectedVerbs) {
        const [moduleName, verbName] = getNames(verbId)
        const module = modules.find((module) => module?.name === moduleName)
        if (module === undefined) {
          continue
        }
        const verb = module?.verbs.find((v) => v.verb?.name === verbName)
        verb && verbs.push([moduleName, verb])
      }
      const calls = await Promise.all(verbs.map(async ([moduleName, verb]) => await getCalls(moduleName, verb?.verb?.name)))
      console.log(calls)
      setCalls(calls.flatMap( call => call))
    }

    fetchCalls()
  }, [client, modules])
  return (
    <Panel className={className}>
      <Panel.Header>Verb Requests(s)</Panel.Header>
      <Panel.Body>
        {
          calls.map((call, index) => <RequestRow key={index} call={call} />)
        }
      </Panel.Body>
    </Panel>
  )
}