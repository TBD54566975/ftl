import { Timestamp } from '@bufbuild/protobuf'
import { createClient } from '../hooks/use-client'
import { ConsoleService } from '../protos/xyz/block/ftl/v1/console/console_connect'
import {
  Call,
  Event,
  EventType,
  EventsQuery_CallFilter,
  EventsQuery_EventTypeFilter,
  EventsQuery_Filter,
  EventsQuery_Order,
  EventsQuery_RequestFilter,
  EventsQuery_TimeFilter,
} from '../protos/xyz/block/ftl/v1/console/console_pb'

const client = createClient(ConsoleService)

const requestKeysFilter = (requestKeys: string[]): EventsQuery_Filter => {
  const filter = new EventsQuery_Filter()
  const requestFilter = new EventsQuery_RequestFilter()
  requestFilter.requests = requestKeys
  filter.filter = {
    case: 'requests',
    value: requestFilter,
  }
  return filter
}

const eventTypesFilter = (eventTypes: EventType[]): EventsQuery_Filter => {
  const filter = new EventsQuery_Filter()
  const typesFilter = new EventsQuery_EventTypeFilter()
  typesFilter.eventTypes = eventTypes
  filter.filter = {
    case: 'eventTypes',
    value: typesFilter,
  }
  return filter
}

const callFilter = (
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

export const getRequestCalls = async (requestKey: string): Promise<Call[]> => {
  const allEvents = await getEvents([requestKeysFilter([requestKey]), eventTypesFilter([EventType.CALL])])
  return allEvents.map((e) => e.entry.value) as Call[]
}

export const getCalls = async (
  destModule: string,
  destVerb: string | undefined = undefined,
  sourceModule: string | undefined = undefined,
): Promise<Call[]> => {
  const allEvents = await getEvents([
    callFilter(destModule, destVerb, sourceModule),
    eventTypesFilter([EventType.CALL]),
  ])
  return allEvents.map((e) => e.entry.value) as Call[]
}

export const getEvents = async (
  filters: EventsQuery_Filter[] = [],
  limit: number = 100,
  order: EventsQuery_Order = EventsQuery_Order.DESC,
): Promise<Event[]> => {
  const response = await client.getEvents({ filters, limit, order })

  return response.events
}
