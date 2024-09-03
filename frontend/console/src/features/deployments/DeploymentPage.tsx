import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useModules } from '../../api/modules/use-modules'
import { modulesFilter } from '../../api/timeline'
import { Badge } from '../../components/Badge'
import { Card } from '../../components/Card'
import { Page } from '../../layout'
import type { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { Timeline } from '../timeline/Timeline'
import { isCron, isExported, isHttpIngress } from '../verbs/verb.utils'

const timeSettings = { isTailing: true, isPaused: false }

export const DeploymentPage = ({ moduleName }: { moduleName: string }) => {
  const modules = useModules()
  const navigate = useNavigate()
  const [module, setModule] = useState<Module | undefined>()

  const filters = useMemo(() => {
    if (!module?.deploymentKey) return []

    return [modulesFilter([module.deploymentKey])]
  }, [module?.deploymentKey])

  useEffect(() => {
    if (modules.isSuccess && modules.data.modules.length > 0) {
      const module = modules.data.modules.find((module) => module.name === moduleName)
      setModule(module)
    }
  }, [modules.data, moduleName])

  return (
    <SidePanelProvider>
      <Page>
        <Page.Body className='flex'>
          <div className=''>
            <div className='flex-1 h-1/2 p-4 overflow-y-scroll'>
              <div className='grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4'>
                {module?.verbs.map((verb) => (
                  <Card key={verb.verb?.name} topBarColor='bg-green-500' onClick={() => navigate(`/modules/${module.name}/verb/${verb.verb?.name}`)}>
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
