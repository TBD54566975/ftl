import { useContext, useMemo, useState } from 'react'
import { useModuleTraceEvents } from '../../api/timeline/use-module-trace-events'
import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../providers/side-panel-provider'
import TimelineEventList from '../timeline/TimelineEventList'
import { TimelineCallDetails } from '../timeline/details/TimelineCallDetails'
import { groupEventsByRequestKey } from './traces.utils'

export const TraceRequestList = ({ module, verb }: { module: string; verb?: string }) => {
  const { openPanel, closePanel } = useContext(SidePanelContext)
  const [selectedEventId, setSelectedEventId] = useState<bigint | undefined>()

  const traceEventsRequest = useModuleTraceEvents(module, verb)
  const traceEvents = useMemo(() => {
    return groupEventsByRequestKey(traceEventsRequest.data)
  }, [traceEventsRequest.data])

  const handleCallClicked = (event: Event) => {
    if (selectedEventId === event.id) {
      setSelectedEventId(undefined)
      closePanel()
      return
    }
    setSelectedEventId(event.id)
    openPanel(<TimelineCallDetails event={event} />)
  }

  const events = Object.entries(traceEvents).map(([_, events]) => events[0])

  return (
    <div className='flex flex-col h-full relative '>
      <TimelineEventList events={events} selectedEventId={selectedEventId} handleEntryClicked={handleCallClicked} />
    </div>
  )
}
