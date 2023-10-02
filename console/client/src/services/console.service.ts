import { Code, ConnectError } from '@bufbuild/connect'
import { Timestamp } from '@bufbuild/protobuf'
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

interface IDFilterParams {
  lowerThan?: bigint
  higherThan?: bigint
}

export const eventIdFilter = ({ lowerThan, higherThan }: IDFilterParams): EventsQuery_Filter => {
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

export const getRequestCalls = async (requestKey: string): Promise<CallEvent[]> => {
  const allEvents = await getEvents({
    filters: [requestKeysFilter([requestKey]), eventTypesFilter([EventType.CALL])],
  })
  return allEvents.map((e) => e.entry.value) as CallEvent[]
}

export const getCalls = async (
  destModule: string,
  destVerb: string | undefined = undefined,
  sourceModule: string | undefined = undefined,
): Promise<CallEvent[]> => {
  const allEvents = await getEvents({
    filters: [callFilter(destModule, destVerb, sourceModule), eventTypesFilter([EventType.CALL])],
  })
  return allEvents.map((e) => e.entry.value) as CallEvent[]
}

interface GetEventsParams {
  limit?: number
  order?: EventsQuery_Order
  filters?: EventsQuery_Filter[]
}

export const getEvents = async ({
  limit = 1000,
  order = EventsQuery_Order.DESC,
  filters = [],
}: GetEventsParams): Promise<Event[]> => {
  const response = await client.getEvents({ filters, limit, order })
  return response.events
}

export interface StreamEventsParams {
  abortControllerSignal: AbortSignal
  filters: EventsQuery_Filter[]
  onEventReceived: (event: Event) => void
}

export const streamEvents = async ({ abortControllerSignal, filters, onEventReceived }: StreamEventsParams) => {
  try {
    for await (const response of client.streamEvents(
      { updateInterval: { seconds: BigInt(1) }, query: { limit: 1000, filters } },
      { signal: abortControllerSignal },
    )) {
      if (response.event != null) {
        onEventReceived(response.event)
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
