import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { useContext, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { CodeBlock } from '../../components/CodeBlock'
import { Page } from '../../layout'
import { CallEvent, EventType, Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { callFilter, eventTypesFilter, streamEvents } from '../../services/console.service'
import { CallList } from '../calls/CallList'
import { VerbForm } from './VerbForm'
import { NotificationType, NotificationsContext } from '../../providers/notifications-provider'

export const VerbPage = () => {
  const { deploymentKey, verbName } = useParams()
  const notification = useContext(NotificationsContext)
  const navgation = useNavigate()
  const modules = useContext(modulesContext)
  const [module, setModule] = useState<Module | undefined>()
  const [verb, setVerb] = useState<Verb | undefined>()
  const [calls, setCalls] = useState<CallEvent[]>([])

  useEffect(() => {
    if (modules.modules.length == 0 || !deploymentKey || !verbName) return

    let module = modules.modules.find((module) => module.deploymentKey === deploymentKey)
    if (!module) {
      const lastIndex = deploymentKey.lastIndexOf('-')
      if (lastIndex !== -1) {
        module = modules.modules.find((module) => module.name === deploymentKey.substring(0, lastIndex))
        navgation(`/deployments/${module?.deploymentKey}/verbs/${verbName}`)
        notification.showNotification({
          title: 'Showing latest deployment',
          message: `The previous deployment of ${module?.deploymentKey} was not found. Showing the latest deployment of ${module?.name}.${verbName} instead.`,
          type: NotificationType.Info,
        })
        setModule(module)
      }
    }
    setModule(module)
    const verb = module?.verbs.find((verb) => verb.verb?.name.toLocaleLowerCase() === verbName?.toLocaleLowerCase())
    setVerb(verb)
  }, [modules, deploymentKey])

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
            { label: module?.deploymentKey || '', link: `/deployments/${module?.deploymentKey}` },
          ]}
        />
        <Page.Body className='p-4'>
          <div className='flex-1 flex flex-col h-full'>
            <div className='flex-1 flex flex-grow h-1/2 mb-4'>
              <div className='mr-2 flex-1 w-1/2 overflow-y-auto'>
                {verb?.verb?.request?.toJsonString() && <CodeBlock code={verb?.schema} language='json' />}
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
