import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useParams } from 'react-router-dom'
import { CodeBlock } from '../../components/CodeBlock'
import { Page } from '../../layout'
import { CallEvent, Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { getCalls } from '../../services/console.service'
import { CallList } from '../calls/CallList'
import { buildVerbSchema } from './verb.utils'

export const VerbPage = () => {
  const { moduleName, verbName } = useParams()
  const modules = React.useContext(modulesContext)
  const [module, setModule] = React.useState<Module | undefined>()
  const [verb, setVerb] = React.useState<Verb | undefined>()
  const [calls, setCalls] = React.useState<CallEvent[] | undefined>()

  const callData =
    module?.data.filter((data) => [verb?.verb?.request?.name, verb?.verb?.response?.name].includes(data.data?.name)) ??
    []

  React.useEffect(() => {
    if (modules) {
      const module = modules.modules.find((module) => module.name === moduleName?.toLocaleLowerCase())
      setModule(module)
      const verb = module?.verbs.find((verb) => verb.verb?.name.toLocaleLowerCase() === verbName?.toLocaleLowerCase())
      setVerb(verb)
    }
  }, [modules, moduleName])

  React.useEffect(() => {
    const abortController = new AbortController()
    if (!module) return

    const fetchCalls = async () => {
      const calls = await getCalls({
        abortControllerSignal: abortController.signal,
        destModule: module.name,
        destVerb: verb?.verb?.name,
      })
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
          title={verb?.verb?.name || ''}
          breadcrumbs={[
            { label: 'Modules', link: '/modules' },
            { label: module?.name || '', link: `/modules/${module?.name}` },
          ]}
        />
        <Page.Body className='p-4'>
          <div className='flex-1 flex flex-col h-full'>
            <div className='flex-1 h-1/2 overflow-y-auto'>
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
            <div className='flex-1 h-1/2'>
              <CallList calls={calls} />
            </div>
          </div>
        </Page.Body>
      </Page>
    </SidePanelProvider>
  )
}
