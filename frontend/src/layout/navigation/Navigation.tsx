import { CubeTransparentIcon, ListBulletIcon, ServerStackIcon, Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { NavLink } from 'react-router-dom'
import { DarkModeSwitch } from '../../components'
import { classNames } from '../../utils'
import { Version } from './Version'

const navigation = [
  { name: 'Events', href: '/events', icon: ListBulletIcon },
  { name: 'Deployments', href: '/deployments', icon: Square3Stack3DIcon },
  { name: 'Modules', href: '/modules', icon: Square3Stack3DIcon },
  { name: 'Graph', href: '/graph', icon: CubeTransparentIcon },
  { name: 'Infrastructure', href: '/infrastructure', icon: ServerStackIcon },
]

export const Navigation = ({ version }: { version?: string }) => {
  return (
    <nav className='bg-indigo-600'>
      <div className='mx-auto pl-3 pr-4'>
        <div className='flex h-16 items-center justify-between'>
          <div className='flex items-center'>
            <div>
              <div className='flex items-baseline space-x-4'>
                {navigation.map((item) => (
                  <NavLink
                    key={item.name}
                    to={item.href}
                    className={({ isActive }) =>
                      classNames(
                        isActive ? 'bg-indigo-700 text-white' : 'text-white hover:bg-indigo-500 hover:bg-opacity-75',
                        'rounded-md px-3 py-2 text-sm font-medium flex items-center space-x-2',
                      )
                    }
                  >
                    <item.icon className='text-lg size-6' />
                    <span className='hidden md:inline'>{item.name}</span>
                  </NavLink>
                ))}
              </div>
            </div>
          </div>
          <div>
            <div className='ml-4 flex items-center space-x-4'>
              <Version version={version} />
              <DarkModeSwitch />
            </div>
          </div>
        </div>
      </div>
    </nav>
  )
}
