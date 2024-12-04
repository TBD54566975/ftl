import { useContext, useMemo, useState } from 'react'
import { useModuleTraceEvents } from '../../api/timeline/use-module-trace-events'
import type { Event } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { SidePanelContext } from '../../providers/side-panel-provider'
import TimelineEventList from '../timeline/TimelineEventList'
import { TimelineCallDetails } from '../timeline/details/TimelineCallDetails'
import { TimelineDetailsHeader } from '../timeline/details/TimelineDetailsHeader'
import { TimelineIngressDetails } from '../timeline/details/TimelineIngressDetails'
import { groupEventsByRequestKey } from './traces.utils'

export const TraceRequestList = ({ module, verb }: { module: string; verb?: string }) => {
  const { openPanel, closePanel } = useContext(SidePanelContext)
  const [selectedEventId, setSelectedEventId] = useState<bigint | undefined>()

  const traceEventsRequest = useModuleTraceEvents(module, verb)
  const traceEvents = useMemo(() => groupEventsByRequestKey(traceEventsRequest.data), [traceEventsRequest.data])

  const handleCallClicked = (event: Event) => {
    if (selectedEventId === event.id) {
      setSelectedEventId(undefined)
      closePanel()
      return
    }
    setSelectedEventId(event.id)
    switch (event.entry?.case) {
      case 'call':
        openPanel(<TimelineCallDetails event={event} />, <TimelineDetailsHeader event={event} />)
        break
      case 'ingress':
        openPanel(<TimelineIngressDetails event={event} />, <TimelineDetailsHeader event={event} />)
        break
    }
  }

  // Get the root/initial event from each request trace group
  const rootEvents = Object.values(traceEvents).map((events) => events[0])

  return (
    <div className='flex flex-col h-full relative '>
      <TimelineEventList events={rootEvents} selectedEventId={selectedEventId} handleEntryClicked={handleCallClicked} />
    </div>
  )
}
