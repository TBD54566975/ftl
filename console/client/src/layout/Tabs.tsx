import {Tab as TabComponent} from '@headlessui/react'
import {XMarkIcon} from '@heroicons/react/24/outline'
import React from 'react'
import {useSearchParams} from 'react-router-dom'
import {Timeline} from '../features/timeline/Timeline'
import {VerbTab} from '../features/verbs/VerbTab'
import {modulesContext} from '../providers/modules-provider'
import {
  NotificationType,
  NotificationsContext,
} from '../providers/notifications-provider'
import {Tab, TabsContext} from '../providers/tabs-provider'
import {headerColor, headerTextColor, panelColor} from '../utils'

const selectedTabStyle = `${headerTextColor} ${headerColor}`
const unselectedTabStyle = `text-gray-300 bg-slate-100 dark:bg-slate-600`

export const Tabs = () => {
  const [searchParams] = useSearchParams()
  const {tabs, activeTabId, openTab, closeTab} = React.useContext(TabsContext)
  const {showNotification} = React.useContext(NotificationsContext)
  const {modules} = React.useContext(modulesContext)

  React.useEffect(() => {
    openTab({
      id: 'timeline',
      label: 'Timeline',
      isClosable: false,
      component: <Timeline />,
    } as Tab)
  }, [])

  React.useEffect(() => {
    const verbRef = searchParams.get('verb')
    if (!verbRef || modules.length === 0) return

    const [moduleName, verbName] = verbRef.split('.')
    const module = modules.find(module => module?.name === moduleName)
    if (!module) {
      showNotification(
        {
          title: 'Module not found',
          message: `Module '${moduleName}' does not exist`,
          type: NotificationType.Error,
        },
        10000
      )
      return
    }

    const verb = module?.verbs.find(v => v.verb?.name === verbName)
    if (!verb) {
      showNotification(
        {
          title: 'Verb not found',
          message: `Verb '${verbName}' does not exist on module '${moduleName}'`,
          type: NotificationType.Error,
        },
        10000
      )
      return
    }

    openTab({
      id: verbRef,
      label: verb.verb?.name ?? 'Verb',
      isClosable: true,
      component: <VerbTab id={verbRef} />,
    })
  }, [modules])

  const handleChangeTab = (index: number) => {
    const tab = tabs[index]
    openTab(tab)
  }

  return (
    <div className={`flex-1 flex flex-col rounded`}>
      <TabComponent.Group
        selectedIndex={tabs.findIndex(tab => tab.id === activeTabId)}
        onChange={handleChangeTab}
      >
        <div>
          <TabComponent.List
            className={`flex items-center rounded-t ${headerTextColor}`}
          >
            {tabs.map(tab => {
              return (
                <TabComponent
                  key={tab.id}
                  className='flex items-center mr-1 relative cursor-pointer'
                  as='span'
                >
                  <span
                    className={`px-4 py-2 rounded-t ${
                      tab.isClosable ? 'pr-8' : ''
                    } ${
                      activeTabId === tab.id
                        ? `${selectedTabStyle}`
                        : `${unselectedTabStyle}`
                    }`}
                  >
                    {tab.label}
                  </span>
                  {tab.isClosable && (
                    <button
                      onClick={e => {
                        e.preventDefault()
                        closeTab(tab.id)
                      }}
                      className='absolute right-0 mr-2 text-gray-400 hover:text-white'
                    >
                      <XMarkIcon className={`h-5 w-5`} />
                    </button>
                  )}
                </TabComponent>
              )
            })}
          </TabComponent.List>
          <div className='flex-grow'></div>
        </div>
        <div className={`flex-1 overflow-y-scroll ${panelColor}`}>
          <TabComponent.Panels>
            {tabs.map(tab => (
              <TabComponent.Panel key={tab.id}>
                {tab.component}
              </TabComponent.Panel>
            ))}
          </TabComponent.Panels>
        </div>
      </TabComponent.Group>
    </div>
  )
}
