import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { useContext, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { CodeBlock } from '../../components/CodeBlock'
import { Page } from '../../layout'
import { CallEvent, EventType, Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { callFilter, eventTypesFilter, streamEvents } from '../../services/console.service'
import { CallList } from '../calls/CallList'
import { VerbForm } from './VerbForm'
import { buildVerbSchema } from './verb.utils'

export const VerbPage = () => {
  const { deploymentName, verbName } = useParams()
  const modules = useContext(modulesContext)
  const [module, setModule] = useState<Module | undefined>()
  const [verb, setVerb] = useState<Verb | undefined>()
  const [calls, setCalls] = useState<CallEvent[]>([])

  const callData =
    module?.data.filter((data) => [verb?.verb?.request?.name, verb?.verb?.response?.name].includes(data.data?.name)) ??
    []

  useEffect(() => {
    if (modules) {
      const module = modules.modules.find((module) => module.deploymentName === deploymentName?.toLocaleLowerCase())
      setModule(module)
      const verb = module?.verbs.find((verb) => verb.verb?.name.toLocaleLowerCase() === verbName?.toLocaleLowerCase())
      setVerb(verb)
    }
  }, [modules, deploymentName])

  useEffect(() => {
    const abortController = new AbortController()
    if (!module) return

    const streamCalls = async () => {
      setCalls([])
      streamEvents({
        abortControllerSignal: abortController.signal,
        filters: [callFilter(module.name, verb?.verb?.name), eventTypesFilter([EventType.CALL])],
        onEventsReceived: (events) => {
          const callEvents = events.map((event) => event.entry.value as CallEvent)
          setCalls((prev) => [...callEvents, ...prev])
        },
      })
    }
    streamCalls()

    return () => {
      abortController.abort()
    }
  }, [module])

  return (
    <SidePanelProvider>
      <Page>
        <Page.Header
          icon={<Square3Stack3DIcon />}
          title={verb?.verb?.name || ''}
          breadcrumbs={[
            { label: 'Deployments', link: '/deployments' },
            { label: module?.deploymentName || '', link: `/deployments/${module?.deploymentName}` },
          ]}
        />
        <Page.Body className='p-4'>
          <div className='flex-1 flex flex-col h-full'>
            <div className='flex-1 flex flex-grow h-1/2 mb-4'>
              <div className='mr-2 flex-1 w-1/2 overflow-y-auto'>
                {verb?.verb?.request?.toJsonString() && (
                  <CodeBlock
                    code={buildVerbSchema(
                      verb?.schema,
                      callData.map((d) => d.schema),
                    )}
                    language='json'
                  />
                )}
              </div>
              <div className='ml-2 flex-1 w-1/2 overflow-y-auto'>
                <VerbForm module={module} verb={verb} />
              </div>
            </div>
            <div className='flex-1 h-1/2'>
              <CallList calls={calls} />
            </div>
          </div>
        </Page.Body>
      </Page>
    </SidePanelProvider>
  )
}
