import type { Timestamp } from '@bufbuild/protobuf'
import {
  type EventType,
  EventsQuery_CallFilter,
  EventsQuery_DeploymentFilter,
  EventsQuery_EventTypeFilter,
  EventsQuery_Filter,
  EventsQuery_IDFilter,
  EventsQuery_LogLevelFilter,
  EventsQuery_RequestFilter,
  EventsQuery_TimeFilter,
  type LogLevel,
} from '../../protos/xyz/block/ftl/v1/console/console_pb'

export const requestKeysFilter = (requestKeys: string[]): EventsQuery_Filter => {
  const filter = new EventsQuery_Filter()
  const requestFilter = new EventsQuery_RequestFilter()
  requestFilter.requests = requestKeys
  filter.filter = {
    case: 'requests',
    value: requestFilter,
  }
  return filter
}

export const eventTypesFilter = (eventTypes: EventType[]): EventsQuery_Filter => {
  const filter = new EventsQuery_Filter()
  const typesFilter = new EventsQuery_EventTypeFilter()
  typesFilter.eventTypes = eventTypes
  filter.filter = {
    case: 'eventTypes',
    value: typesFilter,
  }
  return filter
}

export const logLevelFilter = (logLevel: LogLevel): EventsQuery_Filter => {
  const filter = new EventsQuery_Filter()
  const logFilter = new EventsQuery_LogLevelFilter()
  logFilter.logLevel = logLevel
  filter.filter = {
    case: 'logLevel',
    value: logFilter,
  }
  return filter
}

export const modulesFilter = (modules: string[]): EventsQuery_Filter => {
  const filter = new EventsQuery_Filter()
  const deploymentsFilter = new EventsQuery_DeploymentFilter()
  deploymentsFilter.deployments = modules
  filter.filter = {
    case: 'deployments',
    value: deploymentsFilter,
  }
  return filter
}

export const callFilter = (destModule: string, destVerb: string | undefined = undefined, sourceModule: string | undefined = undefined): EventsQuery_Filter => {
  const filter = new EventsQuery_Filter()
  const callFilter = new EventsQuery_CallFilter()
  callFilter.destModule = destModule
  callFilter.destVerb = destVerb
  callFilter.sourceModule = sourceModule
  filter.filter = {
    case: 'call',
    value: callFilter,
  }
  return filter
}

export const timeFilter = (olderThan: Timestamp | undefined, newerThan: Timestamp | undefined): EventsQuery_Filter => {
  const filter = new EventsQuery_Filter()
  const timeFilter = new EventsQuery_TimeFilter()
  timeFilter.olderThan = olderThan
  timeFilter.newerThan = newerThan
  filter.filter = {
    case: 'time',
    value: timeFilter,
  }
  return filter
}

export const eventIdFilter = ({
  lowerThan,
  higherThan,
}: {
  lowerThan?: bigint
  higherThan?: bigint
}): EventsQuery_Filter => {
  const filter = new EventsQuery_Filter()
  const idFilter = new EventsQuery_IDFilter()
  idFilter.lowerThan = lowerThan
  idFilter.higherThan = higherThan
  filter.filter = {
    case: 'id',
    value: idFilter,
  }
  return filter
}
