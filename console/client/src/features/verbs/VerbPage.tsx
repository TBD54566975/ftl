import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'

export const VerbPage = () => {
  const { moduleName, verbName } = useParams()
  const modules = React.useContext(modulesContext)
  const [module, setModule] = React.useState<Module | undefined>()
  const [verb, setVerb] = React.useState<Verb | undefined>()

  React.useEffect(() => {
    if (modules) {
      const module = modules.modules.find((module) => module.name === moduleName?.toLocaleLowerCase())
      setModule(module)
      const verb = module?.verbs.find((verb) => verb.verb?.name.toLocaleLowerCase() === verbName?.toLocaleLowerCase())
      setVerb(verb)
    }
  }, [modules, moduleName])

  return (
    <>
      <PageHeader
        icon={<Square3Stack3DIcon />}
        title={verb?.verb?.name || ''}
        breadcrumbs={[
          { label: 'Modules', link: '/modules' },
          { label: module?.name || '', link: `/modules/${module?.name}` },
        ]}
      />
      <div className='m-4'>
        <h1>Verb: {verb?.verb?.name}</h1>
      </div>
    </>
  )
}
