import type { Timestamp } from '@bufbuild/protobuf'
import type { EventType, LogLevel } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import {
  GetTimelineRequest_CallFilter,
  GetTimelineRequest_DeploymentFilter,
  GetTimelineRequest_EventTypeFilter,
  GetTimelineRequest_Filter,
  GetTimelineRequest_IDFilter,
  GetTimelineRequest_LogLevelFilter,
  GetTimelineRequest_ModuleFilter,
  GetTimelineRequest_RequestFilter,
  GetTimelineRequest_TimeFilter,
} from '../../protos/xyz/block/ftl/timeline/v1/timeline_pb'

export const requestKeysFilter = (requestKeys: string[]): GetTimelineRequest_Filter => {
  const filter = new GetTimelineRequest_Filter()
  const requestFilter = new GetTimelineRequest_RequestFilter()
  requestFilter.requests = requestKeys
  filter.filter = {
    case: 'requests',
    value: requestFilter,
  }
  return filter
}

export const eventTypesFilter = (eventTypes: EventType[]): GetTimelineRequest_Filter => {
  const filter = new GetTimelineRequest_Filter()
  const typesFilter = new GetTimelineRequest_EventTypeFilter()
  typesFilter.eventTypes = eventTypes
  filter.filter = {
    case: 'eventTypes',
    value: typesFilter,
  }
  return filter
}

export const logLevelFilter = (logLevel: LogLevel): GetTimelineRequest_Filter => {
  const filter = new GetTimelineRequest_Filter()
  const logFilter = new GetTimelineRequest_LogLevelFilter()
  logFilter.logLevel = logLevel
  filter.filter = {
    case: 'logLevel',
    value: logFilter,
  }
  return filter
}

export const modulesFilter = (modules: string[]): GetTimelineRequest_Filter => {
  const filter = new GetTimelineRequest_Filter()
  const deploymentsFilter = new GetTimelineRequest_DeploymentFilter()
  deploymentsFilter.deployments = modules
  filter.filter = {
    case: 'deployments',
    value: deploymentsFilter,
  }
  return filter
}

export const callFilter = (destModule: string, destVerb?: string, sourceModule?: string): GetTimelineRequest_Filter => {
  const filter = new GetTimelineRequest_Filter()
  const callFilter = new GetTimelineRequest_CallFilter()
  callFilter.destModule = destModule
  callFilter.destVerb = destVerb
  callFilter.sourceModule = sourceModule
  filter.filter = {
    case: 'call',
    value: callFilter,
  }
  return filter
}

export const moduleFilter = (module: string, verb?: string): GetTimelineRequest_Filter => {
  const filter = new GetTimelineRequest_Filter()
  const moduleFilter = new GetTimelineRequest_ModuleFilter()
  moduleFilter.module = module
  moduleFilter.verb = verb
  filter.filter = {
    case: 'module',
    value: moduleFilter,
  }
  return filter
}

export const timeFilter = (olderThan: Timestamp | undefined, newerThan: Timestamp | undefined): GetTimelineRequest_Filter => {
  const filter = new GetTimelineRequest_Filter()
  const timeFilter = new GetTimelineRequest_TimeFilter()
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
}): GetTimelineRequest_Filter => {
  const filter = new GetTimelineRequest_Filter()
  const idFilter = new GetTimelineRequest_IDFilter()
  idFilter.lowerThan = lowerThan
  idFilter.higherThan = higherThan
  filter.filter = {
    case: 'id',
    value: idFilter,
  }
  return filter
}
