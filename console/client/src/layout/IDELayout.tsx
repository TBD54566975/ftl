import { XMarkIcon } from '@heroicons/react/24/outline'
import { useContext } from 'react'
import { ModuleDetails } from '../features/modules/ModuleDetails'
import { ModulesList } from '../features/modules/ModulesList'
import { Timeline } from '../features/timeline/Timeline'
import { VerbTab } from '../features/verbs/VerbTab'
import { SelectedModuleContext } from '../providers/selected-module-provider'
import { TabType, TabsContext } from '../providers/tabs-provider'
import { headerColor, headerTextColor, panelColor } from '../utils/style.utils'
import { SidePanel } from './SidePanel'

const selectedTabStyle = `${headerTextColor} ${headerColor}`
const unselectedTabStyle = `text-gray-300 bg-slate-100 dark:bg-slate-600`

export function IDELayout() {
  const { selectedModule } = useContext(SelectedModuleContext)
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

          <div className={`flex-1 overflow-y-scroll ${panelColor}`}>
            {activeTab?.type === TabType.Timeline && <Timeline />}
            {activeTab?.type === TabType.Verb && <VerbTab module={selectedModule} verb={activeTab.verb} />}
          </div>
        </div>

        <SidePanel />
      </div>
    </>
  )
}


