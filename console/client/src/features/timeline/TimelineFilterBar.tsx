import { Popover, Transition } from '@headlessui/react'
import { ChevronDownIcon } from '@heroicons/react/20/solid'
import { Fragment } from 'react'
import { panelColor, textColor } from '../../utils/style.utils'

const filters = [
  {
    id: 'events',
    name: 'Event types',
    options: [
      { value: 'log', label: 'Logs' },
      { value: 'call', label: 'Calls' },
      { value: 'deployment', label: 'Deployments' },
    ],
  },
]

interface Props {
  selectedFilters: string[]
  onFilterChange: (filter: string, checked: boolean) => void
}

export const TimelineFilterBar = ({ selectedFilters, onFilterChange }: Props) => {
  const isOptionChecked = (optionValue: string) => {
    return selectedFilters.includes(optionValue)
  }

  return (
    <>
      <div className={`sticky top-0 z-10 ${panelColor} shadow`}>
        <div className='flex items-center justify-between p-4'>
          <Popover.Group className='hidden sm:flex sm:items-baseline sm:space-x-8'>
            {filters.map((section, sectionIdx) => (
              <Popover
                as='div'
                key={section.name}
                id={`desktop-menu-${sectionIdx}`}
                className='relative inline-block text-left'
              >
                <div>
                  <Popover.Button
                    className={`group inline-flex items-center justify-center text-sm font-medium ${textColor} hover:text-gray-900`}
                  >
                    <span>{section.name}</span>
                    {sectionIdx === 0 && selectedFilters.length > 0 ? (
                      <span className='ml-1.5 rounded bg-gray-200 dark:bg-indigo-600 px-1.5 py-0.5 text-xs font-semibold tabular-nums text-gray-700 dark:text-white'>
                        {selectedFilters.length}
                      </span>
                    ) : null}
                    <ChevronDownIcon
                      className='-mr-1 ml-1 h-5 w-5 flex-shrink-0 text-gray-400 group-hover:text-gray-500'
                      aria-hidden='true'
                    />
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
                    className={`absolute left-0 z-10 mt-2 origin-top-right rounded-md bg-white p-4 shadow-2xl ring-1 ring-black ring-opacity-5 focus:outline-none`}
                  >
                    <form className='space-y-4'>
                      {section.options.map((option, optionIdx) => (
                        <div key={option.value} className='flex items-center'>
                          <input
                            id={`filter-${section.id}-${optionIdx}`}
                            name={`${section.id}[]`}
                            defaultValue={option.value}
                            type='checkbox'
                            className='h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600'
                            checked={isOptionChecked(option.value)}
                            onChange={(e) => onFilterChange(option.value, e.target.checked)}
                          />
                          <label
                            htmlFor={`filter-${section.id}-${optionIdx}`}
                            className='ml-3 whitespace-nowrap pr-6 text-sm font-medium text-gray-900'
                          >
                            {option.label}
                          </label>
                        </div>
                      ))}
                    </form>
                  </Popover.Panel>
                </Transition>
              </Popover>
            ))}
          </Popover.Group>
        </div>
      </div>
    </>
  )
}
