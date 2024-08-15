import { Bars3Icon } from '@heroicons/react/24/outline'
import { useState } from 'react'
import { Outlet } from 'react-router-dom'
import useLocalStorage from '../hooks/use-local-storage'
import { bgColor, textColor } from '../utils'
import { Notification } from './Notification'
import { SidePanel } from './SidePanel'
import MobileNavigation from './navigation/MobileNavigation'
import { Navigation } from './navigation/Navigation'

export const Layout = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [isCollapsed, setIsCollapsed] = useLocalStorage('isNavCollapsed', false)

  return (
    <>
      <div className={`h-screen ${bgColor} ${textColor}`}>
        <div className={`${isMobileMenuOpen ? 'block' : 'hidden'} sm:hidden`}>
          <MobileNavigation onClose={() => setIsMobileMenuOpen(false)} />
        </div>

        <div className={'flex justify-between items-center py-2 px-4 bg-indigo-600 sm:hidden'}>
          <div className='flex shrink-0 items-center rounded-md hover:bg-indigo-700'>
            <span className='text-2xl font-medium text-white'>FTL</span>
            <span className='px-2 text-pink-400 text-2xl font-medium'>∞</span>
          </div>
          <button type='button' title='open' onClick={() => setIsMobileMenuOpen(true)}>
            <Bars3Icon className='h-6 w-6 text-white hover:bg-indigo-700' />
          </button>
        </div>

        <div className={`flex flex-col h-full sm:grid ${isCollapsed ? 'sm:grid-cols-[4rem,1fr]' : 'sm:grid-cols-[13rem,1fr]'}`}>
          <Navigation isCollapsed={isCollapsed} setIsCollapsed={setIsCollapsed} />
          <main className='overflow-hidden flex-1'>
            <Outlet />
          </main>
        </div>
      </div>

      <SidePanel />
      <Notification />
    </>
  )
}
