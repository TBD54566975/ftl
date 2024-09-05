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
        <label htmlFor='search' className='sr-only'>
          Search
        </label>
        <div
          id='command-palette-search'
          className='relative block w-full cursor-pointer rounded-md border border-gray-300 bg-white py-1.5 pl-10 pr-3 text-gray-900 ring-1 ring-inset ring-gray-300/10 placeholder:text-gray-400 sm:text-sm sm:leading-6 hover:border-indigo-500 hover:ring-indigo-500 focus-within:border-indigo-500 focus-within:ring-indigo-600'
          onClick={onFocus}
        >
          <div className='pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3'>
            <Search01Icon aria-hidden='true' className='h-5 w-5 text-gray-400' />
          </div>
          <span className='text-gray-400'>Search</span>
          <div className='absolute inset-y-0 right-0 flex items-center pr-3'>
            <span className='text-gray-400 text-xs bg-gray-100 px-2 py-1 rounded-md'>{shortcutText}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
