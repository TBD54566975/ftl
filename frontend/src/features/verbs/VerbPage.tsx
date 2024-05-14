import { BoltIcon, Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { useContext, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Page } from '../../layout'
import { CallEvent, EventType, Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { callFilter, eventTypesFilter, streamEvents } from '../../services/console.service'
import { NotificationType, NotificationsContext } from '../../providers/notifications-provider'
import { ResizablePanels } from '../../components/ResizablePanels'
import { CallList } from '../calls/CallList'
import { VerbRequestForm } from './VerbRequestForm'
import { verbPanels } from './VerbRightPanel'

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
        const moduleName = deploymentKey.substring(0, lastIndex).replaceAll('dpl-', '')
        module = modules.modules.find((module) => module.name === moduleName)
        navgation(`/deployments/${module?.deploymentKey}/verbs/${verbName}`)
        notification.showNotification({
          title: 'Showing latest deployment',
          message: `The previous deployment of ${module?.deploymentKey} was not found. Showing the latest deployment of ${module?.name}.${verbName} instead.`,
          type: NotificationType.Info,
        })
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

  const header = (
    <div className='flex items-center gap-2 px-2 py-2'>
      <BoltIcon className='h-5 w-5 text-indigo-600' />
      <div className='flex flex-col min-w-0'>Verb</div>
    </div>
  )

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
        <Page.Body className='flex h-full'>
          <ResizablePanels
            mainContent={<VerbRequestForm module={module} verb={verb} />}
            rightPanelHeader={header}
            rightPanelPanels={verbPanels(verb)}
            bottomPanelContent={<CallList calls={calls} />}
          />
        </Page.Body>
      </Page>
    </SidePanelProvider>
  )
}
