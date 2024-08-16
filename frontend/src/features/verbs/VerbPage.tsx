import { BoltIcon, Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { useContext, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useModules } from '../../api/modules/use-modules'
import { useStreamVerbCalls } from '../../api/timeline/stream-verb-calls'
import { Loader } from '../../components/Loader'
import { ResizablePanels } from '../../components/ResizablePanels'
import { Page } from '../../layout'
import type { CallEvent, Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { NotificationType, NotificationsContext } from '../../providers/notifications-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { CallList } from '../calls/CallList'
import { deploymentKeyModuleName } from '../modules/module.utils'
import { VerbRequestForm } from './VerbRequestForm'
import { verbPanels } from './VerbRightPanel'

export const VerbPage = () => {
  const { deploymentKey, verbName } = useParams()
  const notification = useContext(NotificationsContext)
  const navgation = useNavigate()
  const modules = useModules()
  const [module, setModule] = useState<Module | undefined>()
  const [verb, setVerb] = useState<Verb | undefined>()

  useEffect(() => {
    if (!modules.isSuccess) return
    if (modules.data.modules.length === 0 || !deploymentKey || !verbName) return

    let module = modules.data.modules.find((module) => module.deploymentKey === deploymentKey)
    if (!module) {
      const moduleName = deploymentKeyModuleName(deploymentKey)
      if (moduleName) {
        module = modules.data.modules.find((module) => module.name === moduleName)
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
  }, [modules.data, deploymentKey])

  const callEvents = useStreamVerbCalls(module?.name, verb?.verb?.name)
  const calls: CallEvent[] = callEvents.data || []

  if (!module || !verb || callEvents.isLoading) {
    return (
      <div className='flex justify-center items-center min-h-screen'>
        <Loader />
      </div>
    )
  }

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
