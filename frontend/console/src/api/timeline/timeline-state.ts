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
  TAIL = 'tail', // tail=live (default) or tail=paused
  PAST = 'past', // past=15m, on refresh, will be converted to "now-15m" instead of a fixed time range
  AFTER = 'after',
  BEFORE = 'before',
}

export interface PresetTimeRange {
  label: string
  value?: number
}

export const PRESET_TIME_RANGES: Record<string, TimeRange> = {
  tail: { label: 'Live tail' },
  '5m': { label: '5 minutes', value: 5 * 60 * 1000 },
  '15m': { label: '15 minutes', value: 15 * 60 * 1000 },
  '30m': { label: '30 minutes', value: 30 * 60 * 1000 },
  '1h': { label: '1 hour', value: 60 * 60 * 1000 },
  '24h': { label: '24 hours', value: 24 * 60 * 60 * 1000 },
}

type TimeStateLive = {
  kind: 'live'
  paused: boolean
}

// TODO: Track the time when the user paused the timeline, so that it can be converted to a fixed time range for the URL
type TimeStatePast = {
  kind: 'past'
  preset: string
}

type TimeStateRange = {
  kind: 'range'
  // For now let us require both start and end time
  newerThan: Timestamp
  olderThan: Timestamp
}

type TimeState = TimeStateLive | TimeStatePast | TimeStateRange

export type TimelineState = {
  time: TimeState
  // Deployment key
  modules: string[]
  knownModules?: Module[]
  logLevel: LogLevel
  // eventTypes: EventType[] = [EventType.CALL, EventType.LOG, EventType.DEPLOYMENT_CREATED, EventType.DEPLOYMENT_UPDATED]
  eventTypes: EventType[]
  eventId?: bigint
}

export function newTimelineState(params: URLSearchParams, knownModules: Module[] | undefined): TimelineState {
  const state: TimelineState = {
    time: { kind: 'live', paused: false },
    modules: [],
    knownModules,
    logLevel: LogLevel.TRACE,
    eventTypes: [],
    eventId: undefined,
  }

  let isLive = false
  const collectedTimeRange: {
    newerThan?: Timestamp
    olderThan?: Timestamp
  } = {}

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
    } else if (key === UrlKeys.TAIL) {
      isLive = true
      switch (value) {
        case 'live':
          state.time = { kind: 'live', paused: false }
          break
        case 'paused':
          state.time = { kind: 'live', paused: true }
          break
        default:
          console.error(`Unknown tail value: ${value}`)
      }
    } else if (key === UrlKeys.PAST) {
      const preset = PRESET_TIME_RANGES[value]
      if (preset) {
        state.time = { kind: 'live', paused: false }
        collectedTimeRange.newerThan = Timestamp.fromDate(new Date(Date.now() - preset.value))
      } else {
        console.error(`Unknown past value: ${value}`)
      }
    } else if (key === UrlKeys.AFTER) {
      collectedTimeRange.newerThan = Timestamp.fromDate(new Date(value))
    } else if (key === UrlKeys.BEFORE) {
      collectedTimeRange.olderThan = Timestamp.fromDate(new Date(value))
    }
  }

  if (collectedTimeRange.newerThan !== undefined && collectedTimeRange.olderThan !== undefined) {
    state.time = {
      kind: 'range',
      newerThan: collectedTimeRange.newerThan,
      olderThan: collectedTimeRange.olderThan,
    }
  }

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
  if (state.time.kind === 'range') {
    filters.push(timeFilter(state.time.olderThan, state.time.newerThan))
  }
  if (state.time.kind === 'past') {
    const now = new Date()
    const olderThan = Timestamp.fromDate(now)
    const newerThan = Timestamp.fromDate(new Date(now.getTime() - state.time.milliseconds))
    filters.push(timeFilter(olderThan, newerThan))
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

  switch (state.time.kind) {
    case 'live':
      params.set(UrlKeys.TAIL, state.time.paused ? 'paused' : 'live')
      break
    case 'range':
      if (state.time.newerThan) {
        params.set(UrlKeys.AFTER, state.time.newerThan.toDate().toISOString())
      }
      if (state.time.olderThan) {
        params.set(UrlKeys.BEFORE, state.time.olderThan.toDate().toISOString())
      }
      break
    default:
      console.error(`Unknown time kind: ${state.time}`)
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

export function setKnownModules(oldState: TimelineState, modules?: Module[]): TimelineState {
  const state = { ...oldState }
  if (modules) {
    state.knownModules = modules
  }
  return state
}

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

type HistoryRange =
  | { kind: 'past'; preset: string }
  | { kind: 'history'; preset: string }
  | { kind: 'custom'; newerThan: Timestamp; olderThan: Timestamp }
  | { kind: 'live' }
  | { kind: 'other' }

function getPresetForDuration(milliseconds: number): string | undefined {
  for (const [key, range] of Object.entries(PRESET_TIME_RANGES)) {
    if (range.milliseconds === milliseconds) {
      return key
    }
  }
}

// A helper to work out if the given range matches a known preset.
export function getHistoryRange(state: TimelineState): HistoryRange {
  switch (state.time.kind) {
    case 'live':
      return { kind: 'live' }
    case 'past': {
      const preset = PRESET_TIME_RANGES[state.time.preset]
      if (!preset) {
        console.warn('Invalid preset:', state.time.preset)
        return { kind: 'other' }
      }
      return { kind: 'past', preset }
    }
    case 'range':
      if (!(state.time.newerThan && state.time.olderThan)) {
        console.warn('Invalid history range: both newerThan and olderThan must be set:', state.time)
        return { kind: 'other' }
      }

      {
        const newerThan = state.time.newerThan.toDate()
        const olderThan = state.time.olderThan.toDate()
        const ms = olderThan.getTime() - newerThan.getTime()
        const preset = getPresetForDuration(ms)
        if (preset) {
          return { kind: 'preset', key: preset }
        }

        return { kind: 'custom', newerThan: state.time.newerThan, olderThan: state.time.olderThan }
      }
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
