import { Tab } from '@headlessui/react'
import { XMarkIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import { ModuleDetails } from '../features/modules/ModuleDetails'
import { ModulesList } from '../features/modules/ModulesList'
import { RequestModal } from '../features/requests/RequestsModal'
import { Timeline } from '../features/timeline/Timeline'
import { VerbTab } from '../features/verbs/VerbTab'
import { TabSearchParams, TabType, TabsContext } from '../providers/tabs-provider'
import { headerColor, headerTextColor, panelColor } from '../utils/style.utils'
import { SidePanel } from './SidePanel'

const selectedTabStyle = `${headerTextColor} ${headerColor}`
const unselectedTabStyle = `text-gray-300 bg-slate-100 dark:bg-slate-600`

export function IDELayout() {
  const { tabs,activeTab, setActiveTab, setTabs } = React.useContext(TabsContext)
  const navigate = useNavigate()
  const location = useLocation()
  const [ searchParams ] = useSearchParams()

  const handleCloseTab = id => {
    if (activeTab === id && tabs.length > 1) {
      // Set the next available tab as active, if the current active tab is being closed
      const index = tabs.findIndex(tab => tab.id === id)
      setActiveTab(index - 1 )
      searchParams.delete(TabSearchParams.active)
    }
    setTabs(tabs.filter(tab => tab.id !== id))
    navigate({ ...location, search: searchParams.toString() })
  }

  const handleChangeTab = (index: number) => {
    setActiveTab(index)
    index > 0 && tabs.length > 1
      ? searchParams.set(TabSearchParams.active, tabs[index].id)
      : searchParams.delete(TabSearchParams.active)
    navigate({ ...location, search: searchParams.toString() })
  }

  // Handle opening the correct tab on load
  React.useEffect(() => {
    const id = searchParams.get(TabSearchParams.active)
    if(!id) {
      setActiveTab(0)
      return
    }
    const index = tabs.findIndex(tab => tab.id === id)
    if(index >= 0) {
      setActiveTab(activeTab  ?? index)
      return
    }

    const idArr = id.split('.')
    
    if(idArr.length !== 2 ) {
      throw new Error(`invalid reference ${id}`)
    }
    
    const newTab = {
      id,
      label:id[1],
      type: TabType.Verb,
    }
    const nextTabs = [ ...tabs, newTab ]
    setTabs(nextTabs)
    setActiveTab(nextTabs.length - 1)
  }, [ ])

  return (
    <>
      {/* Main Content */}
      <div className='flex flex-grow overflow-hidden'>

        {/* Left Column */}
        <div className='flex flex-col w-1/4 h-full overflow-hidden'>
          {/* Upper Section */}
          <div className={`flex-1 flex flex-col h-1/3 ${panelColor} ml-2 mt-2 rounded`}>
            <div className={`px-4 py-2 rounded-t ${headerTextColor} ${headerColor}`}>Modules</div>
            <div className='flex-1 p-4 overflow-y-auto'>
              <ModulesList />
            </div>
          </div>

          {/* Lower Section */}
          <div className={`flex-1 flex flex-col ${panelColor} ml-2 mt-2 mb-2 rounded`}>
            <div className={`px-4 py-2 rounded-t ${headerTextColor} ${headerColor}`}>Details</div>
            <div className='flex-1 p-4 overflow-y-auto'>
              <ModuleDetails />
            </div>
          </div>
        </div>

        <div className={`flex-1 flex flex-col m-2 rounded`}>
          <Tab.Group
            selectedIndex={activeTab}
            onChange={handleChangeTab}
          >
            <div >
              <Tab.List className={`flex items-center rounded-t ${headerTextColor}`}>
                {tabs.map(({ label, id }, i) => {
                  return (<Tab
                    key={id}
                    className='flex items-center mr-2 relative'
                    as='span'
                  >
                    <span
                      className={`px-4 py-2 rounded-t ${id !== 'timeline' ? 'pr-8' : ''} ${activeTab === i ? `${selectedTabStyle}` : `${unselectedTabStyle}`}`}
                    >
                      {label}
                    </span>
                    {i !== 0 && (<button
                      onClick={e => {
                        e.stopPropagation()
                        handleCloseTab(id)
                        searchParams.get(TabSearchParams.active) === id && searchParams.delete(TabSearchParams.active)
                        navigate({ ...location, search: searchParams.toString() })
                      }}
                      className='absolute right-0 mr-2 text-gray-400 hover:text-white'
                    >
                      <XMarkIcon className={`h-5 w-5`} />
                    </button>)}
                  </Tab>
                  )
                })}
              </Tab.List>
              <div className='flex-grow'></div>
            </div>
            <div className={`flex-1 p-4 overflow-y-scroll ${panelColor}`}>
              <Tab.Panels>
                {tabs.map(({ id }, i) => {
                  return i === 0
                    ? <Tab.Panel key={id}><Timeline /></Tab.Panel>
                    : <Tab.Panel key={id}><VerbTab id={id} /></Tab.Panel>
                })}
              </Tab.Panels>
            </div>
          </Tab.Group>
        </div>
        <SidePanel />
        <RequestModal />
      </div>
    </>
  )
}

