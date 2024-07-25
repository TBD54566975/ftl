import { XMarkIcon } from '@heroicons/react/24/outline'
import { useContext } from 'react'
import { NavLink } from 'react-router-dom'
import { DarkModeSwitch } from '../../components'
import { modulesContext } from '../../providers/modules-provider'
import { classNames } from '../../utils'
import { navigation } from './navigation-items'

const MobileNavigation = ({ onClose }: { onClose: () => void }) => {
  const modules = useContext(modulesContext)

  return (
    <div className='fixed inset-0 z-50 flex flex-col h-full overflow-y-auto bg-indigo-600 text-white'>
      <div className='flex justify-between items-center py-2 px-4 bg-indigo-600'>
        <div className='flex shrink-0 items-center rounded-md'>
          <span className='text-2xl font-medium text-white'>FTL</span>
          <span className='px-2 text-pink-400 text-2xl font-medium'>∞</span>
        </div>
        <button onClick={onClose}>
          <XMarkIcon className='h-6 w-6 text-white hover:bg-indigo-700' />
        </button>
      </div>
      <nav className='flex-1 p-2'>
        <ul>
          {navigation.map((item) => (
            <li key={item.name}>
              <NavLink
                to={item.href}
                className={({ isActive }) =>
                  classNames(
                    isActive ? 'bg-indigo-600 text-white' : 'text-indigo-200 hover:text-white hover:bg-indigo-600',
                  )
                }
              >
                {({ isActive }) => (
                  <div onClick={onClose} className='group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold'>
                    <item.icon
                      className={classNames(
                        isActive ? 'text-white' : 'text-indigo-200 group-hover:text-white',
                        'h-6 w-6 shrink-0',
                      )}
                      aria-hidden='true'
                    />
                    {item.name}
                    {['/modules', '/deployments'].includes(item.href) && (
                      <span
                        className='ml-auto w-9 min-w-max whitespace-nowrap rounded-full bg-indigo-600 px-2.5 py-0.5 text-center text-xs font-medium leading-5 text-white ring-1 ring-inset ring-indigo-500'
                        aria-hidden='true'
                      >
                        {modules.modules.length}
                      </span>
                    )}
                  </div>
                )}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
      <div className='p-4 mt-auto border-t border-indigo-700'>
        <DarkModeSwitch />
      </div>
    </div>
  )
}

export default MobileNavigation
