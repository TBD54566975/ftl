import { RocketLaunchIcon } from '@heroicons/react/24/outline'
import { useContext, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Badge } from '../../components/Badge'
import { Card } from '../../components/Card'
import { Page } from '../../layout'
import type { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { NotificationType, NotificationsContext } from '../../providers/notifications-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { modulesFilter } from '../../services/console.service'
import { deploymentKeyModuleName } from '../modules/module.utils'
import { Timeline } from '../timeline/Timeline'
import { isCron, isExported, isHttpIngress } from '../verbs/verb.utils'

const timeSettings = { isTailing: true, isPaused: false }

export const DeploymentPage = () => {
  const navigate = useNavigate()
  const { deploymentKey } = useParams()
  const modules = useContext(modulesContext)
  const notification = useContext(NotificationsContext)
  const navgation = useNavigate()
  const [module, setModule] = useState<Module | undefined>()

  const filters = useMemo(() => {
    if (!module?.deploymentKey) return []

    return [modulesFilter([module.deploymentKey])]
  }, [module?.deploymentKey])

  useEffect(() => {
    if (modules.modules.length > 0 && deploymentKey) {
      let module = modules.modules.find((module) => module.deploymentKey === deploymentKey)
      if (!module) {
        const moduleName = deploymentKeyModuleName(deploymentKey)
        if (moduleName) {
          module = modules.modules.find((module) => module.name === moduleName)
          navgation(`/deployments/${module?.deploymentKey}`)
          notification.showNotification({
            title: 'Showing latest deployment',
            message: `The previous deployment of ${module?.deploymentKey} was not found. Showing the latest deployment of ${module?.name} instead.`,
            type: NotificationType.Info,
          })
          setModule(module)
        }
      } else {
        setModule(module)
      }
    }
  }, [modules, deploymentKey])

  return (
    <SidePanelProvider>
      <Page>
        <Page.Header icon={<RocketLaunchIcon />} title={module?.deploymentKey || 'Loading...'} breadcrumbs={[{ label: 'Deployments', link: '/deployments' }]} />

        <Page.Body className='flex'>
          <div className=''>
            <div className='flex-1 h-1/2 p-4 overflow-y-scroll'>
              <div className='grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4'>
                {module?.verbs.map((verb) => (
                  <Card
                    key={verb.verb?.name}
                    topBarColor='bg-green-500'
                    onClick={() => navigate(`/deployments/${module.deploymentKey}/verbs/${verb.verb?.name}`)}
                  >
                    <p className='trucate text-sm overflow-hidden'>{verb.verb?.name}</p>
                    {badges(verb).length > 0 && (
                      <div className='pt-1 space-x-1'>
                        {badges(verb).map((badge) => (
                          <Badge key={badge} name={badge} />
                        ))}
                      </div>
                    )}
                  </Card>
                ))}
              </div>
            </div>
            <div className='cursor-col-resize bg-gray-200 dark:bg-gray-700 hover:bg-indigo-600 h-0.5' />
            <div className='flex-1 h-1/2 overflow-y-scroll'>{module?.deploymentKey && <Timeline timeSettings={timeSettings} filters={filters} />}</div>
          </div>
        </Page.Body>
      </Page>
    </SidePanelProvider>
  )
}

const badges = (verb?: Verb) => {
  const all: string[] = []
  if (isHttpIngress(verb)) all.push('http')
  if (isCron(verb)) all.push('cron')
  if (isExported(verb)) all.push('exported')
  return all
}
