import {Tab} from '@headlessui/react'
import {XMarkIcon} from '@heroicons/react/24/outline'
import React from 'react'
import {useSearchParams} from 'react-router-dom'
import {ModuleDetails} from '../features/modules/ModuleDetails'
import {ModulesList} from '../features/modules/ModulesList'
import {Timeline} from '../features/timeline/Timeline'
import {VerbTab} from '../features/verbs/VerbTab'
import {
  NotificationType,
  NotificationsContext,
} from '../providers/notifications-provider'
import {
  TabSearchParams,
  TabType,
  TabsContext,
  timelineTab,
} from '../providers/tabs-provider'
import {
  bgColor,
  headerColor,
  headerTextColor,
  invalidTab,
  panelColor,
  textColor,
} from '../utils'
import {Navigation} from './Navigation'
import {Notification} from './Notification'
import {SidePanel} from './SidePanel'
import {modulesContext} from '../providers/modules-provider'
const selectedTabStyle = `${headerTextColor} ${headerColor}`
const unselectedTabStyle = `text-gray-300 bg-slate-100 dark:bg-slate-600`

export function IDELayout() {
  const {modules} = React.useContext(modulesContext)
  const {tabs, activeTab, setActiveTab, setTabs} = React.useContext(TabsContext)
  const {showNotification} = React.useContext(NotificationsContext)
  const [searchParams, setSearchParams] = useSearchParams()
  const [activeIndex, setActiveIndex] = React.useState(0)
  const id = searchParams.get(TabSearchParams.id) as string
  const type = searchParams.get(TabSearchParams.type) as string

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

  // Handle loading a page with the tab query parameters
  React.useEffect(() => {
    const msg = invalidTab({id, type})
    if (msg) {
      // Default fallback to timeline
      setActiveTab({id: timelineTab.id, type: timelineTab.type})
      if (type === null && id === null) return

      return showNotification({
        title: 'Invalid Tab',
        message: msg,
        type: NotificationType.Error,
      })
    }
    // Handle timeline tab id
    if (id === timelineTab.id) {
      return setActiveTab({id: timelineTab.id, type: timelineTab.type})
    }
    if (modules.length) {
      const ids = id.split('.')
      // Handle edge case where the id contains and invalid module or verb
      if (modules.length) {
        const [moduleId, verbId] = ids
        // Handle Module does not exist
        const moduleExist = modules.find(module => module?.name === moduleId)
        if (!moduleExist) {
          showNotification({
            title: 'Module not found',
            message: `Module '${moduleId}' does not exist`,
            type: NotificationType.Error,
          })
          return setActiveTab({id: timelineTab.id, type: timelineTab.type})
        }
        // Handle Verb does not exists
        const verbExist = moduleExist?.verbs.some(
          ({verb}) => verb?.name === verbId
        )
        if (!verbExist) {
          showNotification({
            title: 'Verb not found',
            message: `Verb '${verbId}' does not exist on module '${moduleId}'`,
            type: NotificationType.Error,
          })
          return setActiveTab({id: timelineTab.id, type: timelineTab.type})
        }
      }
      // Handle if tab is not already in tab list
      if (
        !tabs.some(
          ({id: tabId, type: tabType}) => tabId === id && tabType === type
        )
      ) {
        const newTab = {
          id,
          label: ids[1],
          type: TabType.Verb,
        }
        const nextTabs = [...tabs, newTab]
        setTabs(nextTabs)
        return setActiveTab({id: newTab.id, type: newTab.type})
      }
      // Handle if tab is in tab list
      return setActiveTab({id, type})
    }
  }, [id, type, modules])

  // Set active tab index whenever our activeTab context changes
  React.useEffect(() => {
    const index = tabs.findIndex(tab => tab.id === activeTab?.id)
    setActiveIndex(index)
  }, [activeTab])

  return (
    <>
      <div className={`h-screen flex flex-col ${bgColor} ${textColor}`}>
        <Navigation />

        <div className='flex-grow flex overflow-hidden p-1'>
          {/* Left Column */}
          <aside className={`w-80 flex flex-col`}>
            {/* Top Section */}
            <div
              className={`flex flex-col h-1/2 overflow-hidden rounded-t ${panelColor}`}
            >
              <header className={`px-4 py-2 ${headerTextColor} ${headerColor}`}>
                Modules
              </header>
              <section className={`${panelColor} p-4 overflow-y-auto`}>
                <ModulesList />
              </section>
            </div>

            {/* Bottom Section */}
            <div
              className={`flex flex-col h-1/2 overflow-hidden mt-1 rounded-t ${panelColor}`}
            >
              <header className={`px-4 py-2 ${headerTextColor} ${headerColor}`}>
                Module Details
              </header>
              <section className={`${panelColor} p-4 overflow-y-auto`}>
                <ModuleDetails />
              </section>
            </div>
          </aside>

          {/* Main Content */}
          <main className='flex-grow flex flex-col overflow-hidden pl-1'>
            <section className='flex-grow overflow-y-auto'>
              <div className='flex flex-grow overflow-hidden h-full'>
                <div className={`flex-1 flex flex-col rounded`}>
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
                              className='flex items-center mr-1 relative'
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
            </section>
          </main>
        </div>
      </div>
      <Notification />
    </>
  )
}
