import { Disclosure } from '@headlessui/react'
import { NavLink } from 'react-router-dom'
import { classNames } from '../utils/react.utils'
import DarkModeSwitch from './DarkModeSwitch'

const navigation = [
  { name: 'Graph', href: '/graph' },
]

export function Navigation() {
  return (
    <>
      <Disclosure as='nav'
        className='bg-indigo-600'
      >
        {() => (
          <>
            <div className='mx-auto max-w-7xl px-4 sm:px-6 lg:px-8'>
              <div className='flex h-16 items-center justify-between'>
                <div className='flex items-center'>
                  <div className='flex-shrink-0 flex items-baseline space-x-4'>
                    <NavLink to='/'>
                      <span className='text-indigo-200 text-xl font-medium'>FTL</span>
                      <span className='px-2 text-rose-400 text-2xl font-medium'>∞</span>
                    </NavLink>
                  </div>
                  <div className='md:block'>
                    <div className='ml-10 flex items-baseline space-x-4'>
                      {navigation.map(item => (
                        <NavLink
                          to={item.href}
                          key={item.name}
                          className={({ isActive }) =>
                            classNames(
                              isActive
                                ? 'bg-indigo-700 text-white'
                                : 'text-white hover:bg-indigo-500 hover:bg-opacity-75',
                              'rounded-md px-3 py-2 text-sm font-medium'
                            )
                          }
                        >
                          {item.name}
                        </NavLink>
                      ))}
                    </div>
                  </div>
                </div>
                <div className='md:block'>
                  <DarkModeSwitch />
                </div>
              </div>
            </div>

            <Disclosure.Panel className='md:hidden'>
              <div className='space-y-1 px-2 pb-3 pt-2 sm:px-3'>
                {navigation.map(item => (
                  <Disclosure.Button
                    key={item.name}
                    as={NavLink}
                    to={item.href}
                    className={({ isActive }) =>
                      classNames(
                        isActive ? 'bg-indigo-700 text-white' : 'text-white hover:bg-indigo-500 hover:bg-opacity-75',
                        'block rounded-md px-3 py-2 text-base font-medium'
                      )
                    }
                  >
                    {item.name}
                  </Disclosure.Button>
                ))}
              </div>
            </Disclosure.Panel>
          </>
        )}
      </Disclosure>
    </>
  )
}
