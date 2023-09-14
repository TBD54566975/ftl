import { Listbox, Transition } from '@headlessui/react'
import { BackwardIcon, CheckIcon, ChevronUpDownIcon, ForwardIcon, PlayIcon } from '@heroicons/react/24/outline'
import React, { Fragment } from 'react'
import { bgColor, borderColor, classNames, panelColor, textColor } from '../../../utils'

interface TimeRange {
  label: string
  value: number
}

export const TIME_RANGES: Record<string, TimeRange> = {
  tail: { label: 'Live tail', value: 0 },
  '5m': { label: 'Past 5 minutes', value: 5 * 60 * 1000 },
  '30m': { label: 'Past 30 minutes', value: 30 * 60 * 1000 },
  '1h': { label: 'Past 1 hour', value: 60 * 60 * 1000 },
  '24h': { label: 'Past 24 hours', value: 24 * 60 * 60 * 1000 },
}

export const TimelineTimeControls = () => {
  const [selected, setSelected] = React.useState(TIME_RANGES['tail'])

  return (
    <>
      <div className='flex items-center h-6'>
        <Listbox value={selected} onChange={setSelected}>
          {({ open }) => (
            <>
              <div className='relative w-40 mr-2 -mt-0.5 items-center'>
                <Listbox.Button
                  className={`relative w-full cursor-pointer rounded-md ${bgColor} ${textColor} py-1 pl-3 pr-10 text-xs text-left shadow-sm ring-1 ring-inset ${borderColor} focus:outline-none focus:ring-2 focus:ring-indigo-600`}
                >
                  <span className='block truncate'>{selected.label}</span>
                  <span className='pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2'>
                    <ChevronUpDownIcon className='h-5 w-5 text-gray-400' aria-hidden='true' />
                  </span>
                </Listbox.Button>

                <Transition
                  show={open}
                  as={Fragment}
                  leave='transition ease-in duration-100'
                  leaveFrom='opacity-100'
                  leaveTo='opacity-0'
                >
                  <Listbox.Options
                    className={`absolute z-10 max-h-60 w-full overflow-auto rounded-md ${panelColor} py-1 text-xs shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none`}
                  >
                    {Object.keys(TIME_RANGES).map((key) => {
                      const timeRange = TIME_RANGES[key]
                      return (
                        <Listbox.Option
                          key={key}
                          className={({ active }) =>
                            classNames(
                              active ? 'bg-indigo-600 text-white' : `${textColor}`,
                              'relative cursor-pointer select-none py-2 pl-3 pr-9',
                            )
                          }
                          value={timeRange}
                        >
                          {({ selected, active }) => (
                            <>
                              <span
                                className={classNames(selected ? 'font-semibold' : 'font-normal', 'block truncate')}
                              >
                                {timeRange.label}
                              </span>

                              {selected ? (
                                <span
                                  className={classNames(
                                    active ? 'text-white' : 'text-indigo-600',
                                    'absolute inset-y-0 right-0 flex items-center pr-4',
                                  )}
                                >
                                  <CheckIcon className='h-4 w-4' aria-hidden='true' />
                                </span>
                              ) : null}
                            </>
                          )}
                        </Listbox.Option>
                      )
                    })}
                  </Listbox.Options>
                </Transition>
              </div>
            </>
          )}
        </Listbox>
        <span className={`isolate inline-flex rounded-md shadow-sm h-6 ${textColor} ${bgColor}`}>
          <button
            type='button'
            className={`relative inline-flex items-center rounded-l-md px-3 text-sm font-semibold ring-1 ring-inset ${borderColor} hover:bg-gray-50 dark:hover:bg-gray-700 focus:z-10`}
          >
            <BackwardIcon className='w-4 h-4' />
          </button>
          <button
            type='button'
            className={`relative -ml-px inline-flex items-center px-3 text-sm font-semibold ring-1 ring-inset ${borderColor} hover:bg-gray-50 dark:hover:bg-gray-700 focus:z-10`}
          >
            <PlayIcon className='w-4 h-4' />
          </button>
          <button
            type='button'
            className={`relative -ml-px inline-flex items-center rounded-r-md px-3 text-sm font-semibold ring-1 ring-inset ${borderColor} hover:bg-gray-50 dark:hover:bg-gray-700 focus:z-10`}
          >
            <ForwardIcon className='w-4 h-4' />
          </button>
        </span>
      </div>
    </>
  )
}
