import { Popover, Transition } from '@headlessui/react'
import { ChevronDownIcon } from '@heroicons/react/20/solid'
import { Fragment } from 'react'
import { textColor } from '../../../utils'

interface TimeRange {
  label: string
  value: number
}

export const TIME_RANGES: Record<string, TimeRange> = {
  '5m': { label: 'Last 5 minutes', value: 5 * 60 * 1000 },
  '30m': { label: 'Last 30 minutes', value: 30 * 60 * 1000 },
  '1h': { label: 'Last 1 hour', value: 60 * 60 * 1000 },
  '24h': { label: 'Last 24 hours', value: 24 * 60 * 60 * 1000 },
}

export const TimeFilter = ({
  selectedRange,
  onSelectedRangeChanged,
}: {
  selectedRange: string
  onSelectedRangeChanged: (range: string) => void
}) => {
  return (
    <Popover.Group className='hidden sm:flex sm:items-baseline sm:space-x-8'>
      <Popover as='div' key='log-levels' id={'desktop-menu-log-levels'} className='relative inline-block text-left'>
        <div>
          <Popover.Button className={`group inline-flex items-center justify-center text-sm font-medium ${textColor} hover:text-gray-900`}>
            <span>{TIME_RANGES[selectedRange].label}</span>
            <ChevronDownIcon className='-mr-1 ml-1 h-5 w-5 flex-shrink-0 text-gray-400 group-hover:text-gray-500' aria-hidden='true' />
          </Popover.Button>
        </div>

        <Transition
          as={Fragment}
          enter='transition ease-out duration-100'
          enterFrom='transform opacity-0 scale-95'
          enterTo='transform opacity-100 scale-100'
          leave='transition ease-in duration-75'
          leaveFrom='transform opacity-100 scale-100'
          leaveTo='transform opacity-0 scale-95'
        >
          <Popover.Panel
            className={'absolute right-0 z-10 mt-2 origin-top-right rounded-md bg-white p-4 shadow-2xl ring-1 ring-black ring-opacity-5 focus:outline-none'}
          >
            <form className='space-y-4'>
              {Object.keys(TIME_RANGES).map((key) => (
                <div key={key} className='flex items-center'>
                  <input
                    id={`time-range-${key}`}
                    name={`time-range-${key}`}
                    type='radio'
                    className='h-4 w-4 rounded-full border-gray-300 text-indigo-600 focus:ring-indigo-600'
                    checked={selectedRange === key}
                    onChange={() => onSelectedRangeChanged(key)}
                  />
                  <label htmlFor={`time-range-${key}`} className='ml-3 whitespace-nowrap pr-6 text-sm font-medium text-gray-900'>
                    {TIME_RANGES[key].label}
                  </label>
                </div>
              ))}
            </form>
          </Popover.Panel>
        </Transition>
      </Popover>
    </Popover.Group>
  )
}
