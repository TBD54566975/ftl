import { PhoneIcon, RocketLaunchIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { EventsQuery_Filter, LogLevel } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../../providers/modules-provider'
import { textColor } from '../../../utils'
import { LogLevelBadgeSmall } from '../../logs/LogLevelBadgeSmall'
import { logLevelColor } from '../../logs/log.utils'
import { FilterPanelSection } from './FilterPanelSection'

const EVENT_TYPES: Record<string, string> = {
  call: 'Call',
  log: 'Log',
  deployment: 'Deployment',
}

const EVENT_TYPE_ICON: Record<string, React.ReactNode> = {
  call: <PhoneIcon className='w-4 h-4 text-indigo-400 ml-1' />,
  log: <LogLevelBadgeSmall logLevel={LogLevel.INFO} />,
  deployment: <RocketLaunchIcon className='w-4 h-4 text-indigo-400 ml-1' />,
}

const LOG_LEVELS: Record<number, string> = {
  1: 'Trace',
  5: 'Debug',
  9: 'Info',
  13: 'Warn',
  17: 'Error',
}

interface Props {
  onFiltersChanged: (filters: EventsQuery_Filter[]) => void
}

export const TimelineFilterPanel = ({ onFiltersChanged }: Props) => {
  const modules = React.useContext(modulesContext)
  const [selectedEventTypes, setSelectedEventTypes] = React.useState<string[]>(Object.keys(EVENT_TYPES))
  const [selectedModules, setSelectedModules] = React.useState<string[]>([])
  const [selectedLogLevel, setSelectedLogLevel] = React.useState<number>(1)

  React.useEffect(() => {
    if (selectedModules.length === 0) {
      setSelectedModules(modules.modules.map((module) => module.name))
    }
  }, [modules])

  React.useEffect(() => {
    onFiltersChanged([])
  }, [selectedEventTypes, setSelectedLogLevel, selectedModules])

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

  const handleLogLevelChanged = (logLevel: string) => {
    setSelectedLogLevel(Number(logLevel))
  }

  return (
    <div className='flex-shrink-0 w-52'>
      <div className='w-full'>
        <div className='mx-auto w-full max-w-md p-2'>
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
                  <label
                    htmlFor={`event-type-${key}`}
                    className={`flex justify-between items-center ${textColor} cursor-pointer`}
                  >
                    {EVENT_TYPES[key]}
                    <span>{EVENT_TYPE_ICON[key]}</span>
                  </label>
                </div>
              </div>
            ))}
          </FilterPanelSection>

          <FilterPanelSection title='Log level'>
            {Object.keys(LOG_LEVELS).map((key) => (
              <div key={key} className='relative flex items-start'>
                <div className='flex h-6 items-center'>
                  <input
                    id={`log-level-${key}`}
                    name={`log-level-${key}`}
                    type='checkbox'
                    checked={selectedLogLevel <= Number(key)}
                    onChange={() => handleLogLevelChanged(key)}
                    className='h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer'
                  />
                </div>
                <div className={`ml-2 text-sm leading-6 w-full`}>
                  <label htmlFor={`log-level-${key}`} className={`${logLevelColor[Number(key)]} flex cursor-pointer`}>
                    {LOG_LEVELS[Number(key)]}
                  </label>
                </div>
              </div>
            ))}
          </FilterPanelSection>

          <FilterPanelSection title='Modules'>
            {modules.modules.map((module) => (
              <div key={module.name} className='relative flex items-start'>
                <div className='flex h-6 items-center'>
                  <input
                    id={`module-${module.name}`}
                    name={`module-${module.name}`}
                    type='checkbox'
                    checked={selectedModules.includes(module.name)}
                    onChange={(e) => handleModuleChanged(module.name, e.target.checked)}
                    className='h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer'
                  />
                </div>
                <div className='ml-2 text-sm leading-6 w-full'>
                  <label htmlFor={`module-${module.name}`} className={`${textColor} flex cursor-pointer`}>
                    {module.name}
                  </label>
                </div>
              </div>
            ))}
          </FilterPanelSection>
        </div>
      </div>
    </div>
  )
}
