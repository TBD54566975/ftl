import React from 'react'
import { useSearchParams } from 'react-router-dom'
import { Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { SelectedModuleContext } from '../../providers/selected-module-provider'
import { TabsContext } from '../../providers/tabs-provider'
import { textColor } from '../../utils'
import { VerbTab } from '../verbs/VerbTab'

export const ModuleDetails = () => {
  const modules = React.useContext(modulesContext)
  const { selectedModule, setSelectedModule } = React.useContext(SelectedModuleContext)
  const { openTab } = React.useContext(TabsContext)
  const [searchParams] = useSearchParams()
  const moduleId = searchParams.get('module')
  // When mounting with a valid module in query params set selected module
  React.useEffect(() => {
    if (moduleId) {
      const module = modules.modules.find((module) => module?.name === moduleId)
      module && setSelectedModule(module)
    }
  })

  if (!selectedModule) {
    return (
      <div className='flex-1 p-4 overflow-auto flex items-center justify-center'>
        <span>No module selected</span>
      </div>
    )
  }

  const handleVerbClicked = (verb: Verb) => {
    const verbId = [selectedModule.name, verb.verb?.name].join('.')
    openTab({
      id: verbId,
      label: verb.verb?.name ?? 'Verb',
      isClosable: true,
      component: <VerbTab id={verbId} />,
    })
  }

  return (
    <div className='flex-1 overflow-auto text-sm font-medium text-gray-500 dark:text-gray-400'>
      <div className='flex justify-between'>
        <dt>Name</dt>
        <dd className={`${textColor}`}>{selectedModule.name}</dd>
      </div>
      <div className='flex pt-2 justify-between'>
        <dt>Deployment</dt>
        <dd className={`${textColor}`}>{selectedModule.deploymentName}</dd>
      </div>
      <div className='flex pt-2 justify-between'>
        <dt>Language</dt>
        <dd className={`${textColor}`}>{selectedModule.language}</dd>
      </div>

      <div className='flex pt-4 justify-between'>
        <dt>Verbs</dt>
        <dd className='text-white flex flex-col space-y-2'>
          {selectedModule.verbs.map((verb, index) => (
            <div
              key={index}
              onClick={() => {
                handleVerbClicked(verb)
              }}
              className='rounded bg-indigo-600 px-2 py-1 text-xs font-semibold text-white text-center cursor-pointer
            shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'
            >
              {verb.verb?.name}
            </div>
          ))}
        </dd>
      </div>

      <div className='flex pt-4 justify-between text-right'>
        <dt>Data</dt>
        <dd className='text-white'>
          <ul className='list-none ml-4'>
            {selectedModule.data.map((data, index) => (
              <li key={index} className={`${textColor}`}>
                <code>{data.data?.name}</code>
              </li>
            ))}
          </ul>
        </dd>
      </div>
    </div>
  )
}
