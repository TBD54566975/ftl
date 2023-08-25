import { Tab } from '@headlessui/react'
import { XMarkIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useSearchParams } from 'react-router-dom'
import { ModuleDetails } from '../features/modules/ModuleDetails'
import { ModulesList } from '../features/modules/ModulesList'
import { Timeline } from '../features/timeline/Timeline'
import { VerbTab } from '../features/verbs/VerbTab'
import { TabSearchParams, TabType, TabsContext, timelineTab, isValidTab } from '../providers/tabs-provider'
import { headerColor, headerTextColor, panelColor } from '../utils/style.utils'
import { SidePanel } from './SidePanel'
import { urlSearchParamsToObject } from '../utils'

const selectedTabStyle = `${headerTextColor} ${headerColor}`
const unselectedTabStyle = `text-gray-300 bg-slate-100 dark:bg-slate-600`

export function IDELayout() {
  const { tabs,activeTab, setActiveTab, setTabs } = React.useContext(TabsContext)
  const [ searchParams, setSearchParams ] = useSearchParams()
  const [ activeIndex, setActiveIndex ] = React.useState(0)
  const id = searchParams.get(TabSearchParams.id) as string
  const type = searchParams.get(TabSearchParams.type) as string
  // Set active tab index whenever our activeTab context changes
  React.useEffect(() => {
    const index = tabs.findIndex(tab => tab.id === activeTab?.id)
    setActiveIndex(index)
  }, [ activeTab ])

  const handleCloseTab = (id: string, index: number) => {
    const nextActiveTab = {
      id: tabs[index - 1].id,
      type: tabs[index - 1].type,
    }
    setSearchParams({
      ...urlSearchParamsToObject(searchParams),
      [TabSearchParams.id]:nextActiveTab.id,
      [TabSearchParams.type]: nextActiveTab.type,
    })
    setActiveTab(nextActiveTab)
    setTabs(tabs.filter(tab => tab.id !== id))
  }

  const handleChangeTab = (index: number) => {
    const nextActiveTab = tabs[index]
    setActiveTab({ id: nextActiveTab.id, type: nextActiveTab.type })
    setSearchParams({
      ...urlSearchParamsToObject(searchParams),
      [TabSearchParams.id]:nextActiveTab.id,
      [TabSearchParams.type]: nextActiveTab.type,
    })
  }
  
  React.useEffect(() => {
    if(isValidTab({ id, type })) { // P1 is a valid tab ID and Type
      // P1.1 id and type are not in tab list
      if(!tabs.some(({ id: tabId, type: tabType }) => tabId === id &&  tabType === type)) {
        const verbIdArray = id.split('.')
        const newTab = {
          id,
          label:verbIdArray[1],
          type: TabType.Verb,
        }
        const nextTabs = [ ...tabs, newTab ]
        setTabs(nextTabs)
      }
      // Set active tab id and type
      setActiveTab({ id, type })
    } else { // P2 is an invalid tab ID and type fallback to timeline
      setActiveTab({ id: timelineTab.id, type: timelineTab.type })
    }
  }, [ id, type ])

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
            selectedIndex={activeIndex}
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
                      className={`px-4 py-2 rounded-t ${id !== 'timeline' ? 'pr-8' : ''} ${activeIndex === i ? `${selectedTabStyle}` : `${unselectedTabStyle}`}`}
                    >
                      {label}
                    </span>
                    {i !== 0 && (<button
                      onClick={e => {
                        e.stopPropagation()
                        handleCloseTab(id, i)
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
      </div>
    </>
  )
}

