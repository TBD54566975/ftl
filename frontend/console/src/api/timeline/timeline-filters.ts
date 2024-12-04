import type { Timestamp } from '@bufbuild/protobuf'
import {
  GetEventsRequest_CallFilter,
  GetEventsRequest_DeploymentFilter,
  GetEventsRequest_EventTypeFilter,
  GetEventsRequest_Filter,
  GetEventsRequest_IDFilter,
  GetEventsRequest_LogLevelFilter,
  GetEventsRequest_ModuleFilter,
  GetEventsRequest_RequestFilter,
  GetEventsRequest_TimeFilter,
} from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { type EventType, type LogLevel } from '../../protos/xyz/block/ftl/v1/event_pb'

export const requestKeysFilter = (requestKeys: string[]): GetEventsRequest_Filter => {
  const filter = new GetEventsRequest_Filter()
  const requestFilter = new GetEventsRequest_RequestFilter()
  requestFilter.requests = requestKeys
  filter.filter = {
    case: 'requests',
    value: requestFilter,
  }
  return filter
}

export const eventTypesFilter = (eventTypes: EventType[]): GetEventsRequest_Filter => {
  const filter = new GetEventsRequest_Filter()
  const typesFilter = new GetEventsRequest_EventTypeFilter()
  typesFilter.eventTypes = eventTypes
  filter.filter = {
    case: 'eventTypes',
    value: typesFilter,
  }
  return filter
}

export const logLevelFilter = (logLevel: LogLevel): GetEventsRequest_Filter => {
  const filter = new GetEventsRequest_Filter()
  const logFilter = new GetEventsRequest_LogLevelFilter()
  logFilter.logLevel = logLevel
  filter.filter = {
    case: 'logLevel',
    value: logFilter,
  }
  return filter
}

export const modulesFilter = (modules: string[]): GetEventsRequest_Filter => {
  const filter = new GetEventsRequest_Filter()
  const deploymentsFilter = new GetEventsRequest_DeploymentFilter()
  deploymentsFilter.deployments = modules
  filter.filter = {
    case: 'deployments',
    value: deploymentsFilter,
  }
  return filter
}

export const callFilter = (destModule: string, destVerb?: string, sourceModule?: string): GetEventsRequest_Filter => {
  const filter = new GetEventsRequest_Filter()
  const callFilter = new GetEventsRequest_CallFilter()
  callFilter.destModule = destModule
  callFilter.destVerb = destVerb
  callFilter.sourceModule = sourceModule
  filter.filter = {
    case: 'call',
    value: callFilter,
  }
  return filter
}

export const moduleFilter = (module: string, verb?: string): GetEventsRequest_Filter => {
  const filter = new GetEventsRequest_Filter()
  const moduleFilter = new GetEventsRequest_ModuleFilter()
  moduleFilter.module = module
  moduleFilter.verb = verb
  filter.filter = {
    case: 'module',
    value: moduleFilter,
  }
  return filter
}

export const timeFilter = (olderThan: Timestamp | undefined, newerThan: Timestamp | undefined): GetEventsRequest_Filter => {
  const filter = new GetEventsRequest_Filter()
  const timeFilter = new GetEventsRequest_TimeFilter()
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
}): GetEventsRequest_Filter => {
  const filter = new GetEventsRequest_Filter()
  const idFilter = new GetEventsRequest_IDFilter()
  idFilter.lowerThan = lowerThan
  idFilter.higherThan = higherThan
  filter.filter = {
    case: 'id',
    value: idFilter,
  }
  return filter
}
