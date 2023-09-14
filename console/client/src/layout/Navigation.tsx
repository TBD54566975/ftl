import { CubeTransparentIcon, ListBulletIcon, Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { useContext } from 'react'
import { Link, NavLink } from 'react-router-dom'
import { DarkModeSwitch } from '../components/DarkModeSwitch'
import { modulesContext } from '../providers/modules-provider'
import { classNames } from '../utils'

const navigation = [
  { name: 'Events', href: '/events', icon: ListBulletIcon },
  { name: 'Modules', href: '/modules', icon: Square3Stack3DIcon },
  { name: 'Graph', href: '/graph', icon: CubeTransparentIcon },
]

export const Navigation = () => {
  const modules = useContext(modulesContext)

  return (
    <div className='bg-gray-800 flex-shrink-0 w-52 h-full'>
      <aside className={`flex flex-col h-full`}>
        <div className='flex flex-col h-full overflow-y-auto bg-indigo-700'>
          <div className='flex grow flex-col overflow-y-auto bg-indigo-700 px-4'>
            <Link to='/events'>
              <div className='-mx-2 space-y-1'>
                <div className='flex shrink-0 items-center p-2 rounded-md hover:bg-indigo-700'>
                  <span className='text-2xl font-medium text-white'>FTL</span>
                  <span className='px-2 text-pink-400 text-2xl font-medium'>∞</span>
                </div>
              </div>
            </Link>
            <nav className='flex flex-1 flex-col pt-4'>
              <ul role='list' className='flex flex-1 flex-col gap-y-7'>
                <li>
                  <ul role='list' className='-mx-2 space-y-1'>
                    {navigation.map((item) => (
                      <li key={item.name}>
                        <NavLink
                          to={item.href}
                          className={({ isActive }) =>
                            classNames(
                              isActive
                                ? 'bg-indigo-600 text-white'
                                : 'text-indigo-200 hover:text-white hover:bg-indigo-600',
                              'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold',
                            )
                          }
                        >
                          {({ isActive }) => (
                            <>
                              <item.icon
                                className={classNames(
                                  isActive ? 'text-white' : 'text-indigo-200 group-hover:text-white',
                                  'h-6 w-6 shrink-0',
                                )}
                                aria-hidden='true'
                              />
                              {item.name}
                              {item.href === '/modules' && (
                                <span
                                  className='ml-auto w-9 min-w-max whitespace-nowrap rounded-full bg-indigo-600 px-2.5 py-0.5 text-center text-xs font-medium leading-5 text-white ring-1 ring-inset ring-indigo-500'
                                  aria-hidden='true'
                                >
                                  {modules.modules.length}
                                </span>
                              )}
                            </>
                          )}
                        </NavLink>
                      </li>
                    ))}
                  </ul>
                </li>
                <li className='pb-2 mt-auto'>
                  <DarkModeSwitch />
                </li>
              </ul>
            </nav>
          </div>
        </div>
      </aside>
    </div>
  )
}
