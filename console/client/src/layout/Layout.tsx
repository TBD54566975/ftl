import { Outlet } from 'react-router-dom'
import { Navigation } from './Navigation'
import { Notification } from './Notification'
import { SidePanel } from './SidePanel'

export const Layout = () => {
  return (
    <div className='flex h-screen'>
      <Navigation />

      <main className='flex-1 overflow-y-auto'>
        <section className='h-full relative'>
          <Outlet />
        </section>
      </main>

      <SidePanel />
      <Notification />
    </div>
  )
}
