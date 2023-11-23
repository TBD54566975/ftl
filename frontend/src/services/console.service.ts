import { Timestamp } from '@bufbuild/protobuf'
import { Code, ConnectError } from '@connectrpc/connect'
import { createClient } from '../hooks/use-client'
import { ConsoleService } from '../protos/xyz/block/ftl/v1/console/console_connect'
import {
  CallEvent,
  Event,
  EventType,
  EventsQuery_CallFilter,
  EventsQuery_DeploymentFilter,
  EventsQuery_EventTypeFilter,
  EventsQuery_Filter,
  EventsQuery_IDFilter,
  EventsQuery_LogLevelFilter,
  EventsQuery_Order,
  EventsQuery_RequestFilter,
  EventsQuery_TimeFilter,
  LogLevel,
} from '../protos/xyz/block/ftl/v1/console/console_pb'

const client = createClient(ConsoleService)

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

export const callFilter = (
  destModule: string,
  destVerb: string | undefined = undefined,
  sourceModule: string | undefined = undefined,
): EventsQuery_Filter => {
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

export const getRequestCalls = async ({
  abortControllerSignal,
  requestKey,
}: {
  abortControllerSignal: AbortSignal
  requestKey: string
}): Promise<CallEvent[]> => {
  const allEvents = await getEvents({
    abortControllerSignal,
    filters: [requestKeysFilter([requestKey]), eventTypesFilter([EventType.CALL])],
  })
  return allEvents.map((e) => e.entry.value) as CallEvent[]
}

export const getCalls = async ({
  abortControllerSignal,
  destModule,
  destVerb,
  sourceModule,
}: {
  abortControllerSignal: AbortSignal
  destModule: string
  destVerb?: string
  sourceModule?: string
}): Promise<CallEvent[]> => {
  const allEvents = await getEvents({
    abortControllerSignal,
    filters: [callFilter(destModule, destVerb, sourceModule), eventTypesFilter([EventType.CALL])],
  })
  return allEvents.map((e) => e.entry.value) as CallEvent[]
}

export const getEvents = async ({
  abortControllerSignal,
  limit = 1000,
  order = EventsQuery_Order.DESC,
  filters = [],
}: {
  abortControllerSignal: AbortSignal
  limit?: number
  order?: EventsQuery_Order
  filters?: EventsQuery_Filter[]
}): Promise<Event[]> => {
  const response = await client.getEvents({ filters, limit, order }, { signal: abortControllerSignal })
  return response.events
}

export const streamEvents = async ({
  abortControllerSignal,
  filters,
  onEventsReceived,
}: {
  abortControllerSignal: AbortSignal
  filters: EventsQuery_Filter[]
  onEventsReceived: (events: Event[]) => void
}) => {
  try {
    for await (const response of client.streamEvents(
      { updateInterval: { seconds: BigInt(1) }, query: { limit: 200, filters, order: EventsQuery_Order.DESC } },
      { signal: abortControllerSignal },
    )) {
      if (response.events) {
        onEventsReceived(response.events)
      }
    }
  } catch (error) {
    if (error instanceof ConnectError) {
      if (error.code !== Code.Canceled) {
        console.error('Console service - streamEvents - Connect error:', error)
      }
    } else {
      console.error('Console service - streamEvents:', error)
    }
  }
}
