import { Outlet } from 'react-router-dom'
import { Navigation } from './Navigation'
import { SidePanel } from './SidePanel'

export const Layout = () => {
  return (
    <div className='flex  h-screen'>
      <Navigation />

      <main className='overflow-hidden'>
        <section className='overflow-y-auto h-full'>
          <Outlet />
        </section>
      </main>

      <SidePanel />
    </div>
  )
}
