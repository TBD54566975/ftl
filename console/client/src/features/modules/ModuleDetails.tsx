import { useContext } from 'react'
import { SelectedModuleContext } from '../../providers/selected-module-provider'
import { TabType, TabsContext, TabSearchParams } from '../../providers/tabs-provider'
import { textColor } from '../../utils/style.utils'
import { useSearchParams, useNavigate , useLocation } from 'react-router-dom'

export function ModuleDetails() {
  const { selectedModule } = useContext(SelectedModuleContext)
  const { tabs, setTabs, setActiveTab } = useContext(TabsContext)
  const navigate = useNavigate()
  const location = useLocation()
  const [ searchParams ] = useSearchParams()

  if (!selectedModule) {
    return (
      <div className='flex-1 p-4 overflow-auto flex items-center justify-center'>
        <span>No module selected</span>
      </div>
    )
  }

  const handleVerbClicked = verb => {
    const tabId = [ selectedModule.name, verb.verb?.name ].join('.')
    const index = tabs.findIndex(tab => tab.id === tabId)
    const existingTab = index !== -1
    let newTab
    if(!existingTab) {
      newTab = {
        id: [ selectedModule.name, verb.verb?.name ].join('.'),
        label: verb.verb?.name ?? 'Verb',
        type: TabType.Verb,
      }
      setTabs([ ...tabs, newTab ])
    }
    setActiveTab(existingTab ? index : tabs.length)
    searchParams.set(TabSearchParams.active, newTab?.id ?? tabs[index].id)
    navigate({ ...location, search: searchParams.toString() })
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
              className='rounded bg-indigo-600 px-2 py-1 text-xs font-semibold text-white text-center
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
                <code>{data.name}</code>
              </li>
            ))}
          </ul>
        </dd>
      </div>
    </div>
  )
}
