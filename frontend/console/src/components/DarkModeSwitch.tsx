import { Switch } from '@headlessui/react'

import { Moon02Icon, Sun03Icon } from 'hugeicons-react'
import { useUserPreferences } from '../providers/user-preferences-provider'
import { classNames } from '../utils/react.utils'

export const DarkModeSwitch = () => {
  const { isDarkMode, setDarkMode } = useUserPreferences()

  return (
    <Switch
      checked={isDarkMode}
      onChange={setDarkMode}
      className={
        'bg-indigo-500 relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none'
      }
    >
      <span className='sr-only'>Dark mode toggle</span>
      <span
        aria-hidden='true'
        className={classNames(
          isDarkMode ? 'translate-x-5' : 'translate-x-0',
          'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-gray-50 dark:bg-indigo-600 shadow ring-0 transition duration-200 ease-in-out',
        )}
      >
        {isDarkMode ? <Moon02Icon className='text-white size-5 p-0.5' /> : <Sun03Icon className='text-indigo-500 size-5 p-0.5' />}
      </span>
    </Switch>
  )
}
