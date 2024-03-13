import { RocketLaunchIcon } from '@heroicons/react/24/outline'
import { useContext, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ButtonSmall } from '../../components/ButtonSmall'
import { Card } from '../../components/Card'
import { Page } from '../../layout'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { MetadataCalls, Ref } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { modulesContext } from '../../providers/modules-provider'
import { modulesFilter } from '../../services/console.service'
import { Timeline } from '../timeline/Timeline'
import { verbRefString } from '../verbs/verb.utils'
import { NotificationType, NotificationsContext } from '../../providers/notifications-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'

const timeSettings = { isTailing: true, isPaused: false }

export const DeploymentPage = () => {
  const navigate = useNavigate()
  const { deploymentName } = useParams()
  const modules = useContext(modulesContext)
  const notification = useContext(NotificationsContext)
  const navgation = useNavigate()
  const [module, setModule] = useState<Module | undefined>()
  const [calls, setCalls] = useState<Ref[]>([])

  const filters = useMemo(() => {
    if (!module?.deploymentName) return []

    return [modulesFilter([module.deploymentName])]
  }, [module?.deploymentName])

  useEffect(() => {
    if (modules.modules.length > 0 && deploymentName) {
      let module = modules.modules.find((module) => module.deploymentName === deploymentName)
      if (!module) {
        const lastIndex = deploymentName.lastIndexOf('-')
        if (lastIndex !== -1) {
          module = modules.modules.find((module) => module.name === deploymentName.substring(0, lastIndex))
          navgation(`/deployments/${module?.deploymentName}`)
          notification.showNotification({
            title: 'Showing latest deployment',
            message: `The previous deployment of ${module?.deploymentName} was not found. Showing the latest deployment of ${module?.name} instead.`,
            type: NotificationType.Info,
          })
          setModule(module)
        }
      } else {
        setModule(module)
      }
    }
  }, [modules, deploymentName])

  useEffect(() => {
    if (!module) return

    const verbCalls: Ref[] = []

    const metadata = module.verbs
      .map((v) => v.verb)
      .map((v) => v?.metadata)
      .flat()

    const metadataCalls = metadata
      .filter((metadata) => metadata?.value.case === 'calls')
      .map((metadata) => metadata?.value.value as MetadataCalls)

    const calls = metadataCalls.map((metadata) => metadata?.calls).flat()

    calls.forEach((call) => {
      if (!verbCalls.find((v) => v.name === call.name && v.module === call.module)) {
        verbCalls.push({ name: call.name, module: call.module } as Ref)
      }
    })

    setCalls(Array.from(verbCalls))
  }, [module])

  const handleCallClick = (verb: Ref) => {
    const module = modules?.modules.find((module) => module.name === verb.module)
    if (module) {
      navigate(`/deployments/${module.deploymentName}/verbs/${verb.name}`)
    }
  }

  return (
    <SidePanelProvider>
      <Page>
        <Page.Header
          icon={<RocketLaunchIcon />}
          title={module?.deploymentName || 'Loading...'}
          breadcrumbs={[{ label: 'Deployments', link: '/deployments' }]}
        />

        <Page.Body>
          <div className='flex-1 flex flex-col h-full'>
            <div className='flex-1 h-1/2 mb-4 p-4'>
              <div className='grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4'>
                {module?.verbs.map((verb) => (
                  <Card
                    key={verb.verb?.name}
                    topBarColor='bg-green-500'
                    onClick={() => navigate(`/deployments/${module.deploymentName}/verbs/${verb.verb?.name}`)}
                  >
                    {verb.verb?.name}
                    <p className='text-xs text-gray-400'>{verb.verb?.name}</p>
                  </Card>
                ))}
              </div>
              <h2 className='pt-4'>Calls</h2>
              {calls.length === 0 && <p className='pt-2 text-sm text-gray-400'>Does not call other verbs</p>}
              <ul className='pt-2 flex space-x-2'>
                {calls?.map((verb) => (
                  <li key={`${module?.name}-${verb.module}-${verb.name}`} className='text-xs'>
                    <ButtonSmall onClick={() => handleCallClick(verb)}>{verbRefString(verb)}</ButtonSmall>
                  </li>
                ))}
              </ul>
            </div>
            <div className='flex-1 h-1/2 overflow-y-auto'>
              {module?.deploymentName && <Timeline timeSettings={timeSettings} filters={filters} />}
            </div>
          </div>
        </Page.Body>
      </Page>
    </SidePanelProvider>
  )
}
