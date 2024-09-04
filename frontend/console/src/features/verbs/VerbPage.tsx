import { FunctionIcon } from 'hugeicons-react'
import { useContext, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useModules } from '../../api/modules/use-modules'
import { useStreamVerbCalls } from '../../api/timeline/stream-verb-calls'
import { Loader } from '../../components/Loader'
import { ResizablePanels } from '../../components/ResizablePanels'
import { Page } from '../../layout'
import type { CallEvent, Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { NotificationType, NotificationsContext } from '../../providers/notifications-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { CallList } from '../calls/CallList'
import { VerbRequestForm } from './VerbRequestForm'
import { verbPanels } from './VerbRightPanel'

export const VerbPage = ({ moduleName, declName }: { moduleName: string; declName: string }) => {
  const notification = useContext(NotificationsContext)
  const navgation = useNavigate()
  const modules = useModules()
  const [module, setModule] = useState<Module | undefined>()
  const [verb, setVerb] = useState<Verb | undefined>()

  useEffect(() => {
    if (!modules.isSuccess) return
    if (modules.data.modules.length === 0 || !moduleName || !declName) return

    let module = modules.data.modules.find((module) => module.name === moduleName)
    if (!module && moduleName) {
      module = modules.data.modules.find((module) => module.name === moduleName)
      navgation(`/modules/${module?.name}/verb/${declName}`)
      notification.showNotification({
        title: 'Showing latest deployment',
        message: `The previous deployment of ${module?.name} was not found. Showing the latest deployment of ${module?.name}.${declName} instead.`,
        type: NotificationType.Info,
      })
    }
    setModule(module)
    const verb = module?.verbs.find((verb) => verb.verb?.name.toLocaleLowerCase() === declName?.toLocaleLowerCase())
    setVerb(verb)
  }, [modules.data, moduleName])

  const callEvents = useStreamVerbCalls(module?.name, verb?.verb?.name)
  const calls: CallEvent[] = callEvents.data || []

  if (!module || !verb) {
    return (
      <div className='flex justify-center items-center min-h-screen'>
        <Loader />
      </div>
    )
  }

  const header = (
    <div className='flex items-center gap-2 px-2 py-2'>
      <FunctionIcon className='h-5 w-5 text-indigo-500' />
      <div className='flex flex-col min-w-0'>Verb</div>
    </div>
  )

  return (
    <SidePanelProvider>
      <Page>
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
