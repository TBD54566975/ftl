import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Card } from '../../components/Card'
import { PageHeader } from '../../components/PageHeader'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'

export const ModulePage = () => {
  const navigate = useNavigate()
  const { moduleName } = useParams()
  const modules = React.useContext(modulesContext)
  const [module, setModule] = React.useState<Module | undefined>()

  React.useEffect(() => {
    if (modules) {
      const module = modules.modules.find((module) => module.name === moduleName?.toLocaleLowerCase())
      setModule(module)
    }
  }, [modules, moduleName])

  return (
    <>
      <PageHeader
        icon={<Square3Stack3DIcon />}
        title={module?.name || ''}
        breadcrumbs={[{ label: 'Modules', link: '/modules' }]}
      />
      <div className='m-4'>
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
    </>
  )
}
