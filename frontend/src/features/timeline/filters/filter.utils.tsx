import { PhoneIcon, RocketLaunchIcon } from '@heroicons/react/24/outline'
import { EventType, type EventsQuery_Filter, LogLevel } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { eventTypesFilter, logLevelFilter, modulesFilter } from '../../../api/timeline'
import { LogLevelBadgeSmall } from '../../logs/LogLevelBadgeSmall'

export const EVENT_TYPES: Record<string, EventFilter> = {
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

export const setFilterOnSearchParamObj = (searchParams: URLSearchParams, key: string, selected: string[], defaultLength: string[]) => {
  if (selected.length === defaultLength) {
    searchParams.delete(key)
    return searchParams
  }
  searchParams.set(key, selected)
  return searchParams
}

export const filtersFromSearchParams = (searchParams: URLSearchParams) => {
  const filters: EventsQuery_Filter[] = []

  const eventTypes = searchParams.get('eventTypes')
  if (eventTypes) {
    const selectedTypes = eventTypes.split(',').map((key) => EVENT_TYPES[key].type)
    filters.push(eventTypesFilter(selectedTypes))
  }

  const logLevel = searchParams.get('logLevel')
  if (logLevel) {
    filters.push(logLevelFilter(Number(logLevel)))
  }
  console.log('filtersFromSearchParams', filters, 'from', searchParams.get('logLevel'))

  return filters
}

export const getSelectedEventTypesFromParam = (searchParams: URLSearchParams, defaultValue: string[]) => {
  const eventTypes = searchParams.get('eventTypes')
  if (!eventTypes) return defaultValue
  return eventTypes.split(',')
}

export const getSelectedLogLevelFromParam = (searchParams: URLSearchParams, defaultValue: string[]) => {
  const logLevel = searchParams.get('logLevel')
  return logLevel || defaultValue
}

export const getSelectedEventTypesFromFilter = (filters: EventsQuery_Filter[], defaultValue: string[]) => {
  const eventTypeFilter = filters.find((f) => f.filter.case === 'eventTypes')
  if (!eventTypeFilter) {
    return defaultValue
  }
  return eventTypeFilter.filter.value.eventTypes.map((n) => EventType[n])
}
