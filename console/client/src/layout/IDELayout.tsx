import { ModuleDetails } from '../features/modules/ModuleDetails'
import { ModulesList } from '../features/modules/ModulesList'
import { bgColor, headerColor, headerTextColor, panelColor, textColor } from '../utils'
import { Navigation } from './Navigation'
import { Notification } from './Notification'
import { SidePanel } from './SidePanel'
import { Tabs } from './Tabs'

export const IDELayout = () => {
  return (
    <>
      <div className={`h-screen flex flex-col ${bgColor} ${textColor}`}>
        <Navigation />

        <div className='flex-grow flex overflow-hidden p-1'>
          {/* Left Column */}
          <aside className={`w-80 flex flex-col`}>
            {/* Top Section */}
            <div className={`flex flex-col h-1/2 overflow-hidden rounded-t ${panelColor}`}>
              <header className={`px-4 py-2 ${headerTextColor} ${headerColor}`}>Modules</header>
              <section className={`${panelColor} p-4 overflow-y-auto`}>
                <ModulesList />
              </section>
            </div>

            {/* Bottom Section */}
            <div className={`flex flex-col h-1/2 overflow-hidden mt-1 rounded-t ${panelColor}`}>
              <header className={`px-4 py-2 ${headerTextColor} ${headerColor}`}>Module Details</header>
              <section className={`${panelColor} p-4 overflow-y-auto`}>
                <ModuleDetails />
              </section>
            </div>
          </aside>

          {/* Main Content */}
          <main className='flex-grow flex flex-col overflow-hidden pl-1'>
            <section className='flex-grow overflow-y-auto'>
              <div className='flex flex-grow overflow-hidden h-full'>
                <Tabs />
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
