import { RocketLaunchIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ButtonSmall } from '../../components/ButtonSmall'
import { Card } from '../../components/Card'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { MetadataCalls, VerbRef } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { modulesContext } from '../../providers/modules-provider'
import { verbRefString } from '../verbs/verb.utils'
import { Page } from '../../layout'

export const DeploymentPage = () => {
  const navigate = useNavigate()
  const { deploymentName } = useParams()
  const modules = React.useContext(modulesContext)
  const [module, setModule] = React.useState<Module | undefined>()
  const [calls, setCalls] = React.useState<VerbRef[]>([])

  React.useEffect(() => {
    if (modules) {
      const module = modules.modules.find((module) => module.deploymentName === deploymentName)
      setModule(module)
    }
  }, [modules, deploymentName])

  React.useEffect(() => {
    if (!module) return

    const verbCalls: VerbRef[] = []

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
        verbCalls.push({ name: call.name, module: call.module } as VerbRef)
      }
    })

    setCalls(Array.from(verbCalls))
  }, [modules])

  return (
    <Page>
      <Page.Header
        icon={<RocketLaunchIcon />}
        title={module?.deploymentName || 'Loading...'}
        breadcrumbs={[{ label: 'Deployments', link: '/deployments' }]}
      />

      <Page.Body className='p-4'>
        <div className='grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4'>
          {module?.verbs.map((verb) => (
            <Card
              key={module.deploymentName}
              topBarColor='bg-green-500'
              onClick={() => navigate(`/modules/${module.name}/verbs/${verb.verb?.name}`)}
            >
              {verb.verb?.name}
              <p className='text-xs text-gray-400'>{verb.verb?.name}</p>
            </Card>
          ))}
        </div>
        <h2 className='py-2'>Calls</h2>
        <ul className='space-y-2'>
          {calls?.map((verb) => (
            <li key={`${module?.name}-${verb.module}-${verb.name}`} className='text-xs'>
              <ButtonSmall>{verbRefString(verb)}</ButtonSmall>
            </li>
          ))}
        </ul>
      </Page.Body>
    </Page>
  )
}
