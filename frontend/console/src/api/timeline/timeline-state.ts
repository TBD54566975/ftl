import { Timestamp } from '@bufbuild/protobuf'
// import { EventsQuery_DeploymentFilter, EventsQuery_EventTypeFilter, EventsQuery_Filter, EventsQuery_LogLevelFilter, EventsQuery_TimeFilter, EventType, LogLevel, Module } from "../../protos/xyz/block/ftl/v1/console/console_pb";
import { eventTypesFilter, logLevelFilter, modulesFilter, specificEventIdFilter, timeFilter } from './timeline-filters'
import { TIME_RANGES, TimeSettings, type TimeRange } from '../../features/timeline/filters/TimelineTimeControls'
import { EventsQuery_Filter, EventType, LogLevel, Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
// import { TIME_RANGES, TimeRange, TimeSettings } from "../../features/timeline/filters/TimelineTimeControls";

enum UrlKeys {
  ID = 'id',
  MODULES = 'modules',
  LOG = 'log',
  TYPES = 'types',
  PAUSED = 'paused',
  TAIL = 'tail',
  AFTER = 'after',
  BEFORE = 'before',
}

// spitballing the different time states
// desc       | tail   | paused | olderThan | newerThan
// -----------|--------|--------|-----------|----------
// last 5m    | false  | NA     | now - 5m  | now
// tail pause | true   | true   | NA        | NA
// tailing    | true   | false  | NA        | NA

/* states:
live paused
live tailing
range
*/

// type Range = {
//   olderThan?: Timestamp;
//   newerThan?: Timestamp;
// }

// TODO: type TimeState = Range | Tail;

// type SelectedModules = string[] | 'all'
type SelectedModules = string[]

// Hides the complexity of the URLSearchParams API and protobuf types.
export type TimelineState = {
  isTailing: boolean
  isPaused: boolean
  timeRange: TimeRange
  olderThan?: Timestamp
  newerThan?: Timestamp
  modules: SelectedModules
  knownModules?: Module[]
  logLevel: LogLevel
  // eventTypes: EventType[] = [EventType.CALL, EventType.LOG, EventType.DEPLOYMENT_CREATED, EventType.DEPLOYMENT_UPDATED]
  eventTypes: EventType[]
  eventId?: bigint
}

export function newTimelineState(params: URLSearchParams, knownModules: Module[] | undefined): TimelineState {
  const state: TimelineState = {
    isTailing: true,
    isPaused: false,
    timeRange: TIME_RANGES.tail,
    olderThan: undefined,
    newerThan: undefined,
    modules: [],
    knownModules,
    logLevel: LogLevel.TRACE,
    eventTypes: [],
    eventId: undefined,
  }

  // Quietly ignore invalid values from the user's URL...
  for (const [key, value] of params.entries()) {
    if (key === UrlKeys.ID) {
      state.eventId = BigInt(value)
    } else if (key === UrlKeys.MODULES) {
      state.modules = value.split(',')
    } else if (key === UrlKeys.LOG) {
      const enumValue = logValueToEnum(value)
      if (enumValue) {
        state.logLevel = enumValue
      }
    } else if (key === UrlKeys.TYPES) {
      const types = value
        .split(',')
        .map((type) => eventTypeValueToEnum(type))
        .filter((type) => type !== undefined)
      if (types.length !== 0) {
        state.eventTypes = types
      }
    } else if (key === UrlKeys.PAUSED) {
      state.isPaused = value === '1'
    } else if (key === UrlKeys.TAIL) {
      state.isTailing = value === '1'
    } else if (key === UrlKeys.AFTER) {
      state.olderThan = Timestamp.fromDate(new Date(value))
    } else if (key === UrlKeys.BEFORE) {
      state.newerThan = Timestamp.fromDate(new Date(value))
    }
  }

  // TODO
  // this.timeRange = this.calculateTimeRange();

  // If we're loading a specific event, we don't want to tail.
  //     setSelectedTimeRange(TIME_RANGES['5m'])
  //     setIsTimelinePaused(true)
  //

  return state
}

export function getFilters(state: TimelineState): EventsQuery_Filter[] {
  const filters: EventsQuery_Filter[] = []
  if (state.eventId) {
    filters.push(specificEventIdFilter(state.eventId))
  }
  if (state.modules.length > 0) {
    filters.push(modulesFilter(state.modules))
  }
  if (state.logLevel) {
    filters.push(logLevelFilter(state.logLevel))
  }
  if (state.eventTypes.length > 0) {
    filters.push(eventTypesFilter(state.eventTypes))
  }
  if (state.olderThan || state.newerThan) {
    filters.push(timeFilter(state.olderThan, state.newerThan))
  }
  return filters
}

export function getSearchParams(state: TimelineState): NicerURLSearchParams {
  const params = new NicerURLSearchParams()

  if (state.eventId) {
    params.set(UrlKeys.ID, state.eventId.toString())
  }
  if (state.modules.length > 0) {
    params.set(UrlKeys.MODULES, state.modules.join(','))
  }
  if (state.logLevel !== LogLevel.TRACE) {
    const logString = logEnumToValue(state.logLevel)
    if (logString) {
      params.set(UrlKeys.LOG, logString)
    }
  }
  if (state.eventTypes.length > 0) {
    const eventTypes = state.eventTypes.map((type) => eventTypeEnumToValue(type)).filter((type) => type !== undefined)
    if (eventTypes.length !== 0) {
      params.set(UrlKeys.TYPES, eventTypes.join(','))
    }
  }
  if (state.olderThan) {
    params.set(UrlKeys.AFTER, state.olderThan.toDate().toISOString())
  }
  if (state.newerThan) {
    params.set(UrlKeys.BEFORE, state.newerThan.toDate().toISOString())
  }
  if (state.isPaused) {
    params.set(UrlKeys.PAUSED, '1')
  }
  // Tailing is on by default, so we only need to set it if it's off.
  if (!state.isTailing) {
    params.set(UrlKeys.TAIL, '0')
  }

  console.log('params', params.toString())

  return params
}

export function getModules(state: TimelineState): Module[] {
  return state.knownModules?.filter((module) => state.modules.includes(module.deploymentKey)) || []
}

export function isModuleSelected(state: TimelineState, deploymentKey: string): boolean {
  return state.modules.includes(deploymentKey)
}

export function setTimeSettings(oldState: TimelineState, timeSettings: TimeSettings): TimelineState {
  const state = { ...oldState }
  state.olderThan = timeSettings.olderThan
  state.newerThan = timeSettings.newerThan
  state.isTailing = timeSettings.isTailing
  state.isPaused = timeSettings.isPaused
  return state
}

export function setKnownModules(oldState: TimelineState, modules?: Module[]): TimelineState {
  const state = { ...oldState }
  if (modules) {
    state.knownModules = modules
  }
  return state
}

// function addOrUpdateFilter(oldState: TimelineState, filter: EventsQuery_Filter): TimelineState {
//   const state = { ...oldState }
//   switch (filter.filter.case) {
//     case 'logLevel':
//       state.logLevel = filter.filter.value.logLevel
//       break
//     case 'deployments':
//       state.modules = filter.filter.value.deployments
//       break
//     case 'eventTypes':
//       state.eventTypes = filter.filter.value.eventTypes
//       break
//     case 'time':
//       state.olderThan = filter.filter.value.olderThan
//       state.newerThan = filter.filter.value.newerThan
//       break
//     case 'id':
//       state.eventId = filter.filter.value.higherThan
//       break
//     default:
//       console.error('Unknown filter type while addOrUpdateFilter', filter)
//   }
//   return state
// }

export function isLogLevelSelected(state: TimelineState, level: LogLevel): boolean {
  return state.logLevel <= level
}

function logValueToEnum(value: string): LogLevel | undefined {
  switch (value) {
    case 'trace':
      return LogLevel.TRACE
    case 'debug':
      return LogLevel.DEBUG
    case 'info':
      return LogLevel.INFO
    case 'warn':
      return LogLevel.WARN
    case 'error':
      return LogLevel.ERROR
    default:
      return undefined
  }
}

function logEnumToValue(level: LogLevel): string | undefined {
  switch (level) {
    case LogLevel.TRACE:
      return 'trace'
    case LogLevel.DEBUG:
      return 'debug'
    case LogLevel.INFO:
      return 'info'
    case LogLevel.WARN:
      return 'warn'
    case LogLevel.ERROR:
      return 'error'
    default:
      return undefined
  }
}

export function isEventTypeSelected(state: TimelineState, type: EventType | string): boolean {
  if (typeof type === 'string') {
    console.log('isEventTypeSelected for string', type)
    const enumType = eventTypeValueToEnum(type)
    if (!enumType) {
      console.warn('Unknown event type', type)
      return false
    }

    return state.eventTypes.includes(enumType)
  }

  return state.eventTypes.includes(type)
}

function eventTypeValueToEnum(value: string): EventType | undefined {
  switch (value) {
    case 'log':
      return EventType.LOG
    case 'call':
      return EventType.CALL
    case 'deploymentCreated':
      return EventType.DEPLOYMENT_CREATED
    case 'deploymentUpdated':
      return EventType.DEPLOYMENT_UPDATED
    default:
      return undefined
  }
}

function eventTypeEnumToValue(type: EventType): string | undefined {
  switch (type) {
    case EventType.LOG:
      return 'log'
    case EventType.CALL:
      return 'call'
    case EventType.DEPLOYMENT_CREATED:
      return 'deploymentCreated'
    case EventType.DEPLOYMENT_UPDATED:
      return 'deploymentUpdated'
    default:
      return undefined
  }
}

// A custom URLSearchParams class that removes unnecessary encoding to make the URL more readable for humons.
export class NicerURLSearchParams extends URLSearchParams {
  toString(): string {
    // sort automatically for more predictable URLs
    this.sort()

    let s = super.toString()

    // we don't want to encode commas in the URL, so we replace '%2C' with ','
    s = s.replace(/%2C/g, ',')
    // similar with : and %3A in dates
    s = s.replace(/%3A/g, ':')

    return s
  }
}
