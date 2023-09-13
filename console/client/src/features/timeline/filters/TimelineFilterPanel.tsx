import { Disclosure } from '@headlessui/react'
import { ChevronUpIcon } from '@heroicons/react/20/solid'
import React from 'react'
import { modulesContext } from '../../../providers/modules-provider'
import { textColor } from '../../../utils'

const EVENT_TYPES: Record<string, string> = {
  call: 'Call',
  log: 'Log',
  deployment: 'Deployment',
}

const headerStyles = 'bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600'

export const TimelineFilterPanel = () => {
  const modules = React.useContext(modulesContext)
  const [selectedEventTypes, setSelectedEventTypes] = React.useState<string[]>(Object.keys(EVENT_TYPES))
  const [selectedModules, setSelectedModules] = React.useState<string[]>([])

  React.useEffect(() => {
    if (selectedModules.length === 0) {
      setSelectedModules(modules.modules.map((module) => module.name))
    }
    console.log(modules)
  }, [modules])

  const handleTypeChanged = (eventType: string, checked: boolean) => {
    if (checked) {
      setSelectedEventTypes((prev) => [...prev, eventType])
    } else {
      setSelectedEventTypes((prev) => prev.filter((filter) => filter !== eventType))
    }
  }

  const handleModuleChanged = (moduleName: string, checked: boolean) => {
    if (checked) {
      setSelectedModules((prev) => [...prev, moduleName])
    } else {
      setSelectedModules((prev) => prev.filter((filter) => filter !== moduleName))
    }
  }

  return (
    <div className='flex-shrink-0 w-52'>
      <div className='w-full'>
        <div className='mx-auto w-full max-w-md p-2'>
          <Disclosure defaultOpen={true}>
            {({ open }) => (
              <>
                <Disclosure.Button
                  className={`flex w-full justify-between rounded-md ${headerStyles} py-1 px-2 text-left text-sm font-medium ${textColor} focus:outline-none focus-visible:ring focus-visible:ring-gray-500 focus-visible:ring-opacity-75`}
                >
                  <span>Event types</span>
                  <ChevronUpIcon className={`${open ? 'rotate-180 transform' : ''} h-5 w-5 text-gray-500`} />
                </Disclosure.Button>
                <Disclosure.Panel className={`px-2 py-2 text-sm ${textColor}`}>
                  <fieldset>
                    <legend className='sr-only'>Event types</legend>
                    <div className='space-y-0.5'>
                      {Object.keys(EVENT_TYPES).map((key) => (
                        <div key={key} className='relative flex items-start'>
                          <div className='flex h-6 items-center'>
                            <input
                              id={`event-type-${key}`}
                              name={`event-type-${key}`}
                              type='checkbox'
                              checked={selectedEventTypes.includes(key)}
                              onChange={(e) => handleTypeChanged(key, e.target.checked)}
                              className='h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer'
                            />
                          </div>
                          <div className='ml-2 text-sm leading-6'>
                            <label htmlFor={`event-type-${key}`} className={`${textColor}`}>
                              {EVENT_TYPES[key]}
                            </label>
                          </div>
                        </div>
                      ))}
                    </div>
                  </fieldset>
                </Disclosure.Panel>
              </>
            )}
          </Disclosure>
          <Disclosure as='div' className='mt-2' defaultOpen={true}>
            {({ open }) => (
              <>
                <Disclosure.Button
                  className={`flex w-full justify-between rounded-md ${headerStyles} py-1 px-2 text-left text-sm font-medium ${textColor} hover:bg-gray-200 focus:outline-none focus-visible:ring focus-visible:ring-gray-500 focus-visible:ring-opacity-75`}
                >
                  <span>Modules</span>
                  <ChevronUpIcon className={`${open ? 'rotate-180 transform' : ''} h-5 w-5 text-gray-500`} />
                </Disclosure.Button>
                <Disclosure.Panel className='px-2 py-2 text-sm text-gray-500'>
                  <fieldset>
                    <legend className='sr-only'>Modules</legend>
                    <div className='space-y-0.5'>
                      {modules.modules.map((module) => (
                        <div key={module.name} className='relative flex items-start'>
                          <div className='flex h-6 items-center'>
                            <input
                              id={`module-${module}`}
                              name={`module-${module}`}
                              type='checkbox'
                              checked={selectedModules.includes(module.name)}
                              onChange={(e) => handleModuleChanged(module.name, e.target.checked)}
                              className='h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer'
                            />
                          </div>
                          <div className='ml-2 text-sm leading-6'>
                            <label htmlFor={`module-${module.name}`} className={`${textColor}`}>
                              {module.name}
                            </label>
                          </div>
                        </div>
                      ))}
                    </div>
                  </fieldset>
                </Disclosure.Panel>
              </>
            )}
          </Disclosure>
        </div>
      </div>
    </div>
  )
}
