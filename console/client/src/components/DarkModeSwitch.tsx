import { Switch } from '@headlessui/react'
import { MoonIcon, SunIcon } from '@heroicons/react/20/solid'
import { useEffect } from 'react'
import useLocalStorage from '../hooks/use-local-storage'
import { classNames } from '../utils/react.utils'

export default function DarkModeSwitch() {
  const [ darkMode, setDarkMode ] = useLocalStorage('dark-mode', false)

  useEffect(() => {
    if (darkMode) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }, [ darkMode ])

  return (
    <Switch
      checked={darkMode}
      onChange={setDarkMode}
      className={`bg-indigo-700 relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2`}
    >
      <span className='sr-only'>Dark mode toggle</span>
      <span
        aria-hidden='true'
        className={classNames(
          darkMode ? 'translate-x-5' : 'translate-x-0',
          `pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white dark:bg-indigo-600 shadow ring-0 transition duration-200 ease-in-out`
        )}
      >
        {darkMode ? <MoonIcon className='text-white p-0.5' /> : <SunIcon className='text-indigo-600 p-0.5' />}
      </span>
    </Switch>
  )
}
