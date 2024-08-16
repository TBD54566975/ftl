import { PhoneIcon, RocketLaunchIcon } from '@heroicons/react/24/outline'
import type React from 'react'
import { useEffect, useState } from 'react'
import { useModules } from '../../../api/modules/use-modules'
import { eventTypesFilter, logLevelFilter, modulesFilter } from '../../../api/timeline'
import { EventType, type EventsQuery_Filter, LogLevel } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { textColor } from '../../../utils'
import { LogLevelBadgeSmall } from '../../logs/LogLevelBadgeSmall'
import { logLevelBgColor, logLevelColor, logLevelRingColor } from '../../logs/log.utils'
import { FilterPanelSection } from './FilterPanelSection'

interface EventFilter {
  label: string
  type: EventType
  icon: React.ReactNode
}

const EVENT_TYPES: Record<string, EventFilter> = {
  call: { label: 'Call', type: EventType.CALL, icon: <PhoneIcon className='w-4 h-4 text-indigo-600 ml-1' /> },
  log: { label: 'Log', type: EventType.LOG, icon: <LogLevelBadgeSmall logLevel={LogLevel.INFO} /> },
  deploymentCreated: {
    label: 'Deployment Created',
    type: EventType.DEPLOYMENT_CREATED,
    icon: <RocketLaunchIcon className='w-4 h-4 text-green-500 ml-1' />,
  },
  deploymentUpdated: {
    label: 'Deployment Updated',
    type: EventType.DEPLOYMENT_UPDATED,
    icon: <RocketLaunchIcon className='w-4 h-4 text-indigo-600 ml-1' />,
  },
}

const LOG_LEVELS: Record<number, string> = {
  1: 'Trace',
  5: 'Debug',
  9: 'Info',
  13: 'Warn',
  17: 'Error',
}

export const TimelineFilterPanel = ({
  onFiltersChanged,
}: {
  onFiltersChanged: (filters: EventsQuery_Filter[]) => void
}) => {
  const modules = useModules()
  const [selectedEventTypes, setSelectedEventTypes] = useState<string[]>(Object.keys(EVENT_TYPES))
  const [selectedModules, setSelectedModules] = useState<string[]>([])
  const [previousModules, setPreviousModules] = useState<string[]>([])
  const [selectedLogLevel, setSelectedLogLevel] = useState<number>(1)

  useEffect(() => {
    if (!modules.isSuccess || modules.data.modules.length === 0) {
      return
    }
    const newModules = modules.data.modules.map((module) => module.deploymentKey)
    const addedModules = newModules.filter((name) => !previousModules.includes(name))

    if (addedModules.length > 0) {
      setSelectedModules((prevSelected) => [...prevSelected, ...addedModules])
    }
    setPreviousModules(newModules)
  }, [modules.data])

  useEffect(() => {
    const filter: EventsQuery_Filter[] = []
    if (selectedEventTypes.length !== Object.keys(EVENT_TYPES).length) {
      const selectedTypes = selectedEventTypes.map((key) => EVENT_TYPES[key].type)

      filter.push(eventTypesFilter(selectedTypes))
    }
    if (selectedLogLevel !== LogLevel.TRACE) {
      filter.push(logLevelFilter(selectedLogLevel))
    }

    filter.push(modulesFilter(selectedModules))

    onFiltersChanged(filter)
  }, [selectedEventTypes, selectedLogLevel, selectedModules])

  const handleTypeChanged = (eventType: string, checked: boolean) => {
    if (checked) {
      setSelectedEventTypes((prev) => [...prev, eventType])
    } else {
      setSelectedEventTypes((prev) => prev.filter((filter) => filter !== eventType))
    }
  }

  const handleModuleChanged = (deploymentKey: string, checked: boolean) => {
    if (checked) {
      setSelectedModules((prev) => [...prev, deploymentKey])
    } else {
      setSelectedModules((prev) => prev.filter((filter) => filter !== deploymentKey))
    }
  }

  const handleLogLevelChanged = (logLevel: string) => {
    setSelectedLogLevel(Number(logLevel))
  }

  return (
    <div className='flex-shrink-0 w-52'>
      <div className='w-full'>
        <div className='mx-auto w-full max-w-md pt-2 pl-2 pb-2'>
          <FilterPanelSection title='Event types'>
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
                <div className='ml-2 text-sm leading-6 w-full'>
                  <label htmlFor={`event-type-${key}`} className={`flex justify-between items-center ${textColor} cursor-pointer`}>
                    {EVENT_TYPES[key].label}
                    <span>{EVENT_TYPES[key].icon}</span>
                  </label>
                </div>
              </div>
            ))}
          </FilterPanelSection>

          <FilterPanelSection title='Log level'>
            <ul className='space-y-1'>
              {Object.keys(LOG_LEVELS).map((key) => (
                <li key={key} onClick={() => handleLogLevelChanged(key)} className='relative flex gap-x-2 cursor-pointer'>
                  <div className='relative flex h-5 w-3 flex-none items-center justify-center'>
                    <div
                      className={`${selectedLogLevel <= Number(key) ? 'h-2.5 w-2.5' : 'h-0.5 w-0.5'} ${
                        selectedLogLevel <= Number(key) ? `${logLevelBgColor[Number(key)]} ${logLevelRingColor[Number(key)]}` : 'bg-gray-300 ring-gray-300'
                      } rounded-full ring-1`}
                    />
                  </div>
                  <p className='flex-auto text-sm leading-5 text-gray-500'>
                    <span className={`${logLevelColor[Number(key)]} flex`}>{LOG_LEVELS[Number(key)]}</span>
                  </p>
                </li>
              ))}
            </ul>
          </FilterPanelSection>

          {modules.isSuccess && (
            <FilterPanelSection title='Modules'>
              <div className='relative flex items-center mb-2'>
                <button
                  type='button'
                  onClick={() => setSelectedModules(modules.data.modules.map((module) => module.deploymentKey))}
                  className='text-indigo-600 cursor-pointer hover:text-indigo-500'
                >
                  Select All
                </button>
                <span className='px-1 text-indigo-700'>|</span>
                <button type='button' onClick={() => setSelectedModules([])} className='text-indigo-600 cursor-pointer hover:text-indigo-500'>
                  Deselect All
                </button>
              </div>
              {modules.data.modules.map((module) => (
                <div key={module.deploymentKey} className='relative flex items-start'>
                  <div className='flex h-6 items-center'>
                    <input
                      id={`module-${module.deploymentKey}`}
                      name={`module-${module.deploymentKey}`}
                      type='checkbox'
                      checked={selectedModules.includes(module.deploymentKey)}
                      onChange={(e) => handleModuleChanged(module.deploymentKey, e.target.checked)}
                      className='h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer'
                    />
                  </div>
                  <div className='ml-2 text-sm leading-6 w-full'>
                    <label htmlFor={`module-${module.deploymentKey}`} className={`${textColor} flex cursor-pointer`}>
                      {module.name}
                    </label>
                  </div>
                </div>
              ))}
            </FilterPanelSection>
          )}
        </div>
      </div>
    </div>
  )
}
