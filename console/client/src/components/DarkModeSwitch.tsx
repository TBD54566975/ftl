import {Switch} from '@headlessui/react';
import {MoonIcon, SunIcon} from '@heroicons/react/20/solid';
import {useEffect} from 'react';
import {useDarkMode} from '../providers/dark-mode-provider';
import {classNames} from '../utils/react.utils';

export default function DarkModeSwitch() {
  const {isDarkMode, setDarkMode} = useDarkMode();

  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [isDarkMode]);

  return (
    <Switch
      checked={isDarkMode}
      onChange={setDarkMode}
      className={`bg-indigo-500 relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none`}>
      <span className='sr-only'>Dark mode toggle</span>
      <span
        aria-hidden='true'
        className={classNames(
          isDarkMode ? 'translate-x-5' : 'translate-x-0',
          `pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white dark:bg-indigo-600 shadow ring-0 transition duration-200 ease-in-out`
        )}>
        {isDarkMode ? (
          <MoonIcon className='text-white p-0.5' />
        ) : (
          <SunIcon className='text-indigo-600 p-0.5' />
        )}
      </span>
    </Switch>
  );
}
