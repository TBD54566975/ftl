import { Call02Icon, Rocket01Icon } from 'hugeicons-react'
import type React from 'react'
import { useEffect, useState } from 'react'
import { eventTypesFilter, logLevelFilter, modulesFilter } from '../../../api/timeline'
import { TimelineState, getModules, isEventTypeSelected, isLogLevelSelected, isModuleSelected } from '../../../api/timeline/timeline-state'
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
  call: { label: 'Call', type: EventType.CALL, icon: <Call02Icon className='w-4 h-4 text-indigo-600 ml-1' /> },
  log: { label: 'Log', type: EventType.LOG, icon: <LogLevelBadgeSmall logLevel={LogLevel.INFO} /> },
  deploymentCreated: {
    label: 'Deployment Created',
    type: EventType.DEPLOYMENT_CREATED,
    icon: <Rocket01Icon className='w-4 h-4 text-green-500 ml-1' />,
  },
  deploymentUpdated: {
    label: 'Deployment Updated',
    type: EventType.DEPLOYMENT_UPDATED,
    icon: <Rocket01Icon className='w-4 h-4 text-indigo-600 ml-1' />,
  },
}

const LOG_LEVELS: [LogLevel, string][] = [
  [LogLevel.TRACE, 'Trace'],
  [LogLevel.DEBUG, 'Debug'],
  [LogLevel.INFO, 'Info'],
  [LogLevel.WARN, 'Warn'],
  [LogLevel.ERROR, 'Error'],
]

export const TimelineFilterPanel = ({
  timelineState,
  setTimelineState,
}: {
  timelineState: TimelineState
  setTimelineState: (state: TimelineState) => void
}) => {
  // const modules = timelineState.knownModules

  // const [selectedEventTypes, setSelectedEventTypes] = useState<string[]>(timelineState.eventTypes.map(eventTypeToKey))
  // const [selectedModules, setSelectedModules] = useState<string[]>(timelineState.getModules())
  // const [selectedLogLevel, setSelectedLogLevel] = useState<number>(timelineState.logLevel)

  // useEffect(() => {
  //   console.log('modules', JSON.stringify(modules))
  //   // if (!modules.isSuccess || modules.data.modules.length === 0) {
  //   //   return
  //   // }
  //   // const newModules = modules.data.modules.map((module) => module.deploymentKey)
  //   // const addedModules = newModules.filter((name) => !previousModules.includes(name))

  //   // if (addedModules.length > 0) {
  //   //   setSelectedModules((prevSelected) => [...prevSelected, ...addedModules])
  //   // }
  //   // setPreviousModules(newModules)

  //   // modules.data are possible modules to choose from
  //   //

  // }, [modules.data])

  // useEffect(() => {
  //   const filter: EventsQuery_Filter[] = []
  //   if (selectedEventTypes.length !== Object.keys(EVENT_TYPES).length) {
  //     console.log('selectedEventTypes', JSON.stringify(selectedEventTypes))
  //     const selectedTypes = selectedEventTypes.map((key) => EVENT_TYPES[key].type)

  //     filter.push(eventTypesFilter(selectedTypes))
  //   }
  //   if (selectedLogLevel !== LogLevel.TRACE) {
  //     filter.push(logLevelFilter(selectedLogLevel))
  //   }

  //   filter.push(modulesFilter(selectedModules))

  //   onFiltersChanged(filter)
  // }, [selectedEventTypes, selectedLogLevel, selectedModules])

  const handleTypeChanged = (eventType: string, checked: boolean) => {
    console.log('TODO handleTypeChanged', eventType, checked)
    if (checked) {
      setTimelineState({ ...timelineState, eventTypes: [...timelineState.eventTypes, EVENT_TYPES[eventType].type] })
    } else {
      setTimelineState({
        ...timelineState,
        eventTypes: timelineState.eventTypes.filter((type) => type !== EVENT_TYPES[eventType].type),
      })
    }
  }

  const handleLogLevelChanged = (logLevel: LogLevel) => {
    setTimelineState({ ...timelineState, logLevel })
  }

  const handleModuleChanged = (deploymentKey: string, checked: boolean) => {
    console.log('TODO handleModuleChanged', deploymentKey, checked)
    // if (checked) {
    //   setSelectedModules((prev) => [...prev, deploymentKey])
    // } else {
    //   setSelectedModules((prev) => prev.filter((filter) => filter !== deploymentKey))
    // }

    if (checked) {
      setTimelineState({ ...timelineState, modules: [...timelineState.modules, deploymentKey] })
    } else {
      setTimelineState({
        ...timelineState,
        modules: timelineState.modules.filter((module) => module !== deploymentKey),
      })
    }
  }

  const selectAllModules = () => {
    setTimelineState({ ...timelineState, modules: getModules(timelineState).map((module) => module.deploymentKey) })
  }

  const clearSelectedModules = () => {
    setTimelineState({ ...timelineState, modules: [] })
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
                    checked={isEventTypeSelected(timelineState, key)}
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
              {LOG_LEVELS.map(([level, label]) => (
                <li key={level} onClick={() => handleLogLevelChanged(level)} className='relative flex gap-x-2 cursor-pointer'>
                  <div className='relative flex h-5 w-3 flex-none items-center justify-center'>
                    <div
                      className={`${isLogLevelSelected(timelineState, level) ? 'h-2.5 w-2.5' : 'h-0.5 w-0.5'} ${
                        isLogLevelSelected(timelineState, level) ? `${logLevelBgColor[Number(level)]} ${logLevelRingColor[level]}` : 'bg-gray-300 ring-gray-300'
                      } rounded-full ring-1`}
                    />
                  </div>
                  <p className='flex-auto text-sm leading-5 text-gray-500'>
                    <span className={`${logLevelColor[level]} flex`}>{label}</span>
                  </p>
                </li>
              ))}
            </ul>
          </FilterPanelSection>

          <FilterPanelSection title='Modules' loading={timelineState.knownModules === undefined}>
            <div className='relative flex items-center mb-2'>
              <button type='button' onClick={selectAllModules} className='text-indigo-600 cursor-pointer hover:text-indigo-500'>
                Select All
              </button>
              <span className='px-1 text-indigo-700'>|</span>
              <button type='button' onClick={clearSelectedModules} className='text-indigo-600 cursor-pointer hover:text-indigo-500'>
                Deselect All
              </button>
            </div>
            {timelineState.knownModules?.map((module) => (
              <div key={module.deploymentKey} className='relative flex items-start'>
                <div className='flex h-6 items-center'>
                  <input
                    id={`module-${module.deploymentKey}`}
                    name={`module-${module.deploymentKey}`}
                    type='checkbox'
                    checked={isModuleSelected(timelineState, module.deploymentKey)}
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
        </div>
      </div>
    </div>
  )
}
