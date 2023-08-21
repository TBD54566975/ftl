import { useContext, useState } from 'react'
import { ModuleDetails } from '../features/modules/ModuleDetails'
import { ModulesList } from '../features/modules/ModulesList'
import { Timeline } from '../features/timeline/Timeline'
import { TimelineEventDetails } from '../features/timeline/TimelineEventDetails'
import { SelectedModuleContext } from '../providers/selected-module-provider'
import { headerColor, headerTextColor, panelColor, textColor } from '../utils/style.utils'
import { SelectedTimelineEntryContext } from '../providers/selected-timeline-entry-provider'
import { XMarkIcon } from '@heroicons/react/24/outline'

const selectedTabStyle = `${headerTextColor} ${headerColor}`
const unselectedTabStyle = `text-gray-300 bg-slate-100 dark:bg-slate-600`

export function IDELayout() {
  const { selectedModule } = useContext(SelectedModuleContext)
  const { selectedEntry, setSelectedEntry } = useContext(SelectedTimelineEntryContext)
  const [ activeTab, setActiveTab ] = useState('timeline')

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

        <div className={`flex-1 flex flex-col  m-2 rounded`}>
          <div className={`flex items-center rounded-t ${headerTextColor}`}>
            <button
              onClick={() => setActiveTab('timeline')}
              className={`mr-2 px-4 py-2 rounded-t ${activeTab === 'timeline' ? `${selectedTabStyle}` : `${unselectedTabStyle}`}`}
            >
              Timeline
            </button>
            <button
              onClick={() => setActiveTab('verb')}
              className={`mr-2 px-4 py-2 rounded-t ${activeTab === 'verb' ? `${selectedTabStyle}` : `${unselectedTabStyle}`}`}
            >
              Verb
            </button>

            <div className='flex-grow'></div>
          </div>


          <div className={`flex-1 p-4 overflow-y-scroll ${panelColor}`}>
            {activeTab === 'timeline' && <Timeline module={selectedModule} />}
            {activeTab === 'verb' && <div>Content for the other tab.</div>}
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


