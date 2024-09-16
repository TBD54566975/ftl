import { Search01Icon } from 'hugeicons-react'
import type React from 'react'
import { useEffect } from 'react'

type SearchInputProps = {
  onFocus: () => void
}

export const SearchInput: React.FC<SearchInputProps> = ({ onFocus }) => {
  const shortcutText = window.navigator.userAgent.includes('Mac') ? '⌘ + K' : 'Ctrl + K'

  useEffect(() => {
    const handleKeydown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
        event.preventDefault()
        onFocus()
      }
    }

    window.addEventListener('keydown', handleKeydown)

    return () => {
      window.removeEventListener('keydown', handleKeydown)
    }
  }, [onFocus])

  return (
    <div className='flex flex-1 items-center justify-center px-2 lg:ml-6 lg:justify-end'>
      <div className='w-full max-w-lg lg:max-w-xs'>
        <div
          id='command-palette-search'
          className='relative block w-full cursor-pointer rounded-md border border-indigo-700 bg-indigo-700/50 py-1.5 pl-10 pr-3 text-indigo-200  placeholder:text-indigo-300 sm:text-sm sm:leading-6 hover:border-indigo-500 hover:ring-indigo-500 focus-within:border-indigo-400 focus-within:ring-indigo-400'
          onClick={onFocus}
        >
          <div className='pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3'>
            <Search01Icon aria-hidden='true' className='h-5 w-5 text-indigo-300' />
          </div>
          <span className='text-indigo-300'>Search</span>
          <div className='absolute inset-y-0 right-0 flex items-center pr-3'>
            <span className='text-indigo-200 text-xs bg-indigo-600 px-2 py-1 rounded-md'>{shortcutText}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
