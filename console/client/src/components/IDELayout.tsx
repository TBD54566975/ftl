import { useContext } from 'react'
import { ModuleDetails } from '../features/modules/ModuleDetails'
import { ModulesList } from '../features/modules/ModulesList'
import { Timeline } from '../features/timeline/Timeline'
import { TimelineEventDetails } from '../features/timeline/TimelineEventDetails'
import { SelectedModuleContext } from '../providers/selected-module-provider'
import { headerColor, headerTextColor, panelColor, textColor } from '../utils/style.utils'
import { SelectedTimelineEntryContext } from '../providers/selected-timeline-entry-provider'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { TabType, TabsContext } from '../providers/tabs-provider'
import { VerbTab } from '../features/verbs/VerbTab'

const selectedTabStyle = `${headerTextColor} ${headerColor}`
const unselectedTabStyle = `text-gray-300 bg-slate-100 dark:bg-slate-600`

export function IDELayout() {
  const { selectedModule } = useContext(SelectedModuleContext)
  const { selectedEntry, setSelectedEntry } = useContext(SelectedTimelineEntryContext)
  const { tabs,activeTab, setActiveTab, setTabs } = useContext(TabsContext)

  const handleCloseTab = id => {
    if (activeTab === id && tabs.length > 1) {
      // Set the next available tab as active, if the current active tab is being closed
      const nextTab = tabs.find(tab => tab.id !== id) || tabs[0]
      setActiveTab(nextTab)
    }
    setTabs(tabs.filter(tab => tab.id !== id))
  }

  return (
    <>
      {/* Main Content */}
      <div className='flex flex-grow overflow-hidden'>

        {/* Left Column */}
        <div className='flex flex-col w-1/4 h-full overflow-hidden'>
          {/* Upper Section */}
          <div className={`flex-1 flex flex-col ${panelColor} ml-2 mt-2 rounded`}>
            <div className={`px-4 py-2 rounded-t ${headerTextColor} ${headerColor}`}>Modules</div>
            <div className='flex-1 p-4 overflow-y-auto'>
              <ModulesList />
            </div>
          </div>

          {/* Lower Section */}
          <div className={`flex-1 flex flex-col max-h-80 ${panelColor} ml-2 mt-2 mb-2 rounded`}>
            <div className={`px-4 py-2 rounded-t ${headerTextColor} ${headerColor}`}>Details</div>
            <div className='flex-1 p-4 overflow-y-auto'>
              <ModuleDetails />
            </div>
          </div>
        </div>

        <div className={`flex-1 flex flex-col m-2 rounded`}>
          <div className={`flex items-center rounded-t ${headerTextColor}`}>
            {tabs.map(tab => (
              <div key={tab.id} className='flex items-center mr-2 relative'>
                <button
                  onClick={() => setActiveTab(tab)}
                  className={`px-4 py-2 rounded-t ${tab.id !== 'timeline' ? 'pr-8' : ''} ${activeTab === tab ? `${selectedTabStyle}` : `${unselectedTabStyle}`}`}
                >
                  {tab.label}
                </button>
                {tab.id !== 'timeline' && (
                  <button
                    onClick={() => handleCloseTab(tab.id)}
                    className='absolute right-0 mr-2 text-gray-400 hover:text-white'
                  >
                    <XMarkIcon className={`h-5 w-5`} />
                  </button>
                )}
              </div>
            ))}
            <div className='flex-grow'></div>
          </div>

          <div className={`flex-1 p-4 overflow-y-scroll ${panelColor}`}>
            {activeTab?.type === TabType.Timeline && <Timeline module={selectedModule} />}
            {activeTab?.type === TabType.Verb && <VerbTab module={selectedModule} verb={activeTab.verb} />}
          </div>
        </div>

        {/* Right Column */}
        <div className={`w-1/3 p-4 flex flex-col bg-white dark:bg-slate-800
               absolute top-0 right-0 h-full transform
               ${selectedEntry != null ? 'translate-x-0 shadow-xl' : 'translate-x-full'}
               transition-transform duration-300`}
        >
          <div className={`rounded-t ${textColor} flex justify-between items-center`}>
            Event Details
            <button onClick={() => setSelectedEntry(null)}
              className='p-1 hover:bg-indigo-100 dark:hover:bg-indigo-500'
            >
              <XMarkIcon className={`h-5 w-5 hover:text-gray-600`} />
            </button>
          </div>
          <div className='w-full h-px bg-gray-300 mt-2'></div>
          <div className='flex-1 pt-2 overflow-auto'>
            <TimelineEventDetails />
          </div>
        </div>
      </div>
    </>
  )
}


