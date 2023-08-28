import React from 'react'
import {Tab} from '@headlessui/react'
import {XMarkIcon} from '@heroicons/react/24/outline'
import {InformationCircleIcon} from '@heroicons/react/20/solid'
import {useSearchParams} from 'react-router-dom'
import {ModuleDetails} from '../features/modules/ModuleDetails'
import {ModulesList} from '../features/modules/ModulesList'
import {Timeline} from '../features/timeline/Timeline'
import {VerbTab} from '../features/verbs/VerbTab'
import {
  TabSearchParams,
  TabType,
  TabsContext,
  timelineTab,
} from '../providers/tabs-provider'
import {headerColor, headerTextColor, panelColor, invalidTab} from '../utils'
import {SidePanel} from './SidePanel'
import {Notification} from '../components/Notification'
import {useClient} from '../hooks/use-client'
import {ConsoleService} from '../protos/xyz/block/ftl/v1/console/console_connect'
const selectedTabStyle = `${headerTextColor} ${headerColor}`
const unselectedTabStyle = `text-gray-300 bg-slate-100 dark:bg-slate-600`

export function IDELayout() {
  const client = useClient(ConsoleService)
  const {tabs, activeTab, setActiveTab, setTabs} = React.useContext(TabsContext)
  const [searchParams, setSearchParams] = useSearchParams()
  const [activeIndex, setActiveIndex] = React.useState(0)
  const [invalidTabMessage, setInvalidTabMessage] = React.useState<string>()
  const id = searchParams.get(TabSearchParams.id) as string
  const type = searchParams.get(TabSearchParams.type) as string
  // Set active tab index whenever our activeTab context changes
  React.useEffect(() => {
    const index = tabs.findIndex(tab => tab.id === activeTab?.id)
    setActiveIndex(index)
  }, [activeTab])

  const handleCloseTab = (id: string, index: number) => {
    const nextActiveTab = {
      id: tabs[index - 1].id,
      type: tabs[index - 1].type,
    }
    setSearchParams({
      ...Object.fromEntries(searchParams),
      [TabSearchParams.id]: nextActiveTab.id,
      [TabSearchParams.type]: nextActiveTab.type,
    })
    setActiveTab(nextActiveTab)
    setTabs(tabs.filter(tab => tab.id !== id))
  }

  const handleChangeTab = (index: number) => {
    const nextActiveTab = tabs[index]
    setActiveTab({id: nextActiveTab.id, type: nextActiveTab.type})
    setSearchParams({
      ...Object.fromEntries(searchParams),
      [TabSearchParams.id]: nextActiveTab.id,
      [TabSearchParams.type]: nextActiveTab.type,
    })
  }

  React.useEffect(() => {
    const validateTabs = async () => {
      const modules = await client.getModules({})
      const msg = invalidTab({id, type})
      if (msg) {
        // IDs an invalid tab ID and type fallback to timeline
        setActiveTab({id: timelineTab.id, type: timelineTab.type})
        // On intial mount we have no query params set for tabs so we want to skip setting invalidTabMessage
        if (type === null && id === null) return
        return setInvalidTabMessage(msg)
      }
      const inTabsList = tabs.some(
        ({id: tabId, type: tabType}) => tabId === id && tabType === type
      )
      // Tab is in tab list just set active tab
      if (inTabsList) return setActiveTab({id, type})
      // Get module and Verb ids
      const [moduleId, verbId] = id.split('.')
      // Check to see if they exist on controller
      const moduleExist = modules.modules.find(
        module => module?.name === moduleId
      )
      const verbExist = moduleExist?.verbs.some(
        ({verb}) => verb?.name === verbId
      )
      // Set tab if they both exists
      if (moduleExist && verbExist) {
        const newTab = {
          id: moduleId,
          label: verbId,
          type: TabType.Verb,
        }
        const nextTabs = [...tabs, newTab]
        setActiveTab({id, type})
        return setTabs(nextTabs)
      }
      if (moduleExist && !verbExist) {
        setInvalidTabMessage(`Verb ${verbId} does not exist on ${moduleId}`)
      }
      if (!moduleExist) {
        setInvalidTabMessage(`Module ${moduleId} does not exist`)
      }
      setActiveTab({id: timelineTab.id, type: timelineTab.type})
    }
    void validateTabs()
  }, [id, type])

  return (
    <>
      {/* Main Content */}
      <div className='flex flex-grow overflow-hidden'>
        {/* Left Column */}
        <div className='flex flex-col w-1/4 h-full overflow-hidden'>
          {/* Upper Section */}
          <div
            className={`flex-1 flex flex-col h-1/3 ${panelColor} ml-2 mt-2 rounded`}
          >
            <div
              className={`px-4 py-2 rounded-t ${headerTextColor} ${headerColor}`}
            >
              Modules
            </div>
            <div className='flex-1 p-4 overflow-y-auto'>
              <ModulesList />
            </div>
          </div>

          {/* Lower Section */}
          <div
            className={`flex-1 flex flex-col ${panelColor} ml-2 mt-2 mb-2 rounded`}
          >
            <div
              className={`px-4 py-2 rounded-t ${headerTextColor} ${headerColor}`}
            >
              Module Details
            </div>
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
            <div>
              <Tab.List
                className={`flex items-center rounded-t ${headerTextColor}`}
              >
                {tabs.map(({label, id}, i) => {
                  return (
                    <Tab
                      key={id}
                      className='flex items-center mr-2 relative'
                      as='span'
                    >
                      <span
                        className={`px-4 py-2 rounded-t ${
                          id !== 'timeline' ? 'pr-8' : ''
                        } ${
                          activeIndex === i
                            ? `${selectedTabStyle}`
                            : `${unselectedTabStyle}`
                        }`}
                      >
                        {label}
                      </span>
                      {i !== 0 && (
                        <button
                          onClick={e => {
                            e.stopPropagation()
                            handleCloseTab(id, i)
                          }}
                          className='absolute right-0 mr-2 text-gray-400 hover:text-white'
                        >
                          <XMarkIcon className={`h-5 w-5`} />
                        </button>
                      )}
                    </Tab>
                  )
                })}
              </Tab.List>
              <div className='flex-grow'></div>
            </div>
            <div className={`flex-1 overflow-y-scroll ${panelColor}`}>
              <Tab.Panels>
                {tabs.map(({id}, i) => {
                  return i === 0 ? (
                    <Tab.Panel key={id}>
                      <Timeline />
                    </Tab.Panel>
                  ) : (
                    <Tab.Panel key={id}>
                      <VerbTab id={id} />
                    </Tab.Panel>
                  )
                })}
              </Tab.Panels>
            </div>
          </Tab.Group>
        </div>
        <SidePanel />
      </div>
      {invalidTabMessage && (
        <Notification
          color='text-red-400'
          icon={
            <InformationCircleIcon className='flex-shrink-0 inline w-4 h-4 mr-3' />
          }
          title='Alert!'
          message={invalidTabMessage}
        />
      )}
    </>
  )
}
