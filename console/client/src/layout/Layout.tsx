import { Outlet } from 'react-router-dom'
import { Navigation } from './Navigation'
import { Notification } from './Notification'
import { SidePanel } from './SidePanel'
import { bgColor, textColor } from '../utils'

export const Layout = () => {
  return (
    <>
      <div
        className={`grid h-screen min-w-[1024px] min-h-[600px] ${bgColor} ${textColor}`}
        style={{ gridTemplateColumns: '13rem 1fr', gridTemplateRows: '100vh'}}
      >
        <Navigation />
        <main className='overflow-hidden'>
          <Outlet />
        </main>
      </div>
      <SidePanel/>
      <Notification />
    </>
    
  )
}
