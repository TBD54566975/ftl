import { NavLink } from 'react-router-dom'
import { DarkModeSwitch } from '../components/DarkModeSwitch'
import { classNames } from '../utils/react.utils'
import { navColor } from '../utils/style.utils'

export const Navigation = () => {
  return (
    <div className={`px-4 py-2 flex items-center justify-between ${navColor} text-white shadow-md`}>
      <div className='flex items-center space-x-2'>
        <NavLink to='/'>
          <div className='pb-1'>
            <span className='text-xl font-medium'>FTL</span>
            <span className='px-2 text-pink-400 text-2xl font-medium'>∞</span>
          </div>
        </NavLink>
        <div className='hidden md:block'>
          <div className='ml-2 flex items-baseline space-x-4'>
            <NavLink
              to='/graph'
              key='graph'
              className={({ isActive }) =>
                classNames(
                  isActive ? 'bg-indigo-700 text-white' : 'text-white hover:bg-indigo-500 hover:bg-opacity-75',
                  'rounded-md px-3 py-2 text-sm font-medium',
                )
              }
            >
              Graph
            </NavLink>
          </div>
        </div>
      </div>
      <DarkModeSwitch />
    </div>
  )
}
