import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Card } from '../../components/Card'
import { Page } from '../../layout'
import { CallEvent, Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { getCalls } from '../../services/console.service'
import { CallList } from '../calls/CallList'

export const ModulePage = () => {
  const navigate = useNavigate()
  const { moduleName } = useParams()
  const modules = React.useContext(modulesContext)
  const [module, setModule] = React.useState<Module | undefined>()
  const [calls, setCalls] = React.useState<CallEvent[] | undefined>()

  React.useEffect(() => {
    if (modules) {
      const module = modules.modules.find((module) => module.name === moduleName?.toLocaleLowerCase())
      setModule(module)
    }
  }, [modules, moduleName])

  React.useEffect(() => {
    const abortController = new AbortController()
    if (!module) return

    const fetchCalls = async () => {
      const calls = await getCalls({ abortControllerSignal: abortController.signal, destModule: module.name })
      setCalls(calls)
    }
    fetchCalls()

    return () => {
      abortController.abort()
    }
  }, [module])

  return (
    <SidePanelProvider>
      <Page>
        <Page.Header
          icon={<Square3Stack3DIcon />}
          title={module?.name || ''}
          breadcrumbs={[{ label: 'Modules', link: '/modules' }]}
        />
        <Page.Body className='p-4'>
          <div className='flex-1 flex flex-col h-full'>
            <div className='flex-1'>
              <div className='grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4'>
                {module?.verbs.map((verb) => (
                  <Card
                    key={verb.verb?.name}
                    topBarColor='bg-green-500'
                    onClick={() => navigate(`/modules/${module.name}/verbs/${verb.verb?.name}`)}
                  >
                    {verb.verb?.name}
                    <p className='text-xs text-gray-400'>{verb.verb?.name}</p>
                  </Card>
                ))}
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
