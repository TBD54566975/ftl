import { useContext } from 'react'
import { ModuleDetails } from '../features/modules/ModuleDetails'
import { ModulesList } from '../features/modules/ModulesList'
import { Timeline } from '../features/timeline/Timeline'
import { TimelineEventDetails } from '../features/timeline/TimelineEventDetails'
import { SelectedModuleContext } from '../providers/selected-module-provider'
import { headerColor, headerTextColor, panelColor } from '../utils/style.utils'

export function IDELayout() {
  const { selectedModule } = useContext(SelectedModuleContext)

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

        {/* Center Column */}
        <div className={`flex-1 flex flex-col ${panelColor} m-2 rounded`}>
          <div className={`px-4 py-2 rounded-t ${headerTextColor} ${headerColor}`}>Timeline</div>
          <div className='flex-1 p-4 overflow-y-scroll'>
            <Timeline module={selectedModule} />
          </div>
        </div>

        {/* Right Column */}
        <div className={`w-1/4 flex flex-col ${panelColor} mr-2 mt-2 mb-2 rounded`}>
          <div className={`px-4 py-2 rounded-t ${headerTextColor} ${headerColor}`}>Event Details</div>
          <div className='flex-1 p-4 overflow-auto'>
            <TimelineEventDetails />
          </div>
        </div>
      </div>
    </>
  )
}


