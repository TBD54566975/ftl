import { Bars3Icon } from '@heroicons/react/24/outline'
import { useState } from 'react'
import { Outlet } from 'react-router-dom'
import { bgColor, textColor } from '../utils'
import { Notification } from './Notification'
import { SidePanel } from './SidePanel'
import MobileNavigation from './navigation/MobileNavigation'
import { Navigation } from './navigation/Navigation'

export const Layout = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)

  return (
    <>
      <div className={`${isMobileMenuOpen ? 'block' : 'hidden'} sm:hidden`}>
        <MobileNavigation onClose={() => setIsMobileMenuOpen(false)} />
      </div>

      <div
        className={`grid h-screen ${bgColor} ${textColor} sm:grid-cols-[13rem,1fr] sm:grid-rows-[100vh] hidden sm:grid`}
      >
        <Navigation />
        <main className='overflow-hidden'>
          <Outlet />
        </main>
      </div>

      <div className={`h-screen ${bgColor} ${textColor} grid sm:hidden`} style={{ gridTemplateRows: 'auto 1fr' }}>
        <div className='flex justify-between items-center py-2 px-4 bg-indigo-600'>
          <div className='flex shrink-0 items-center rounded-md hover:bg-indigo-700'>
            <span className='text-2xl font-medium text-white'>FTL</span>
            <span className='px-2 text-pink-400 text-2xl font-medium'>∞</span>
          </div>
          <button onClick={() => setIsMobileMenuOpen(true)}>
            <Bars3Icon className='h-6 w-6 text-white hover:bg-indigo-700' />
          </button>
        </div>
        <main className='overflow-hidden'>
          <Outlet />
        </main>
      </div>

      <SidePanel />
      <Notification />
    </>
  )
}
