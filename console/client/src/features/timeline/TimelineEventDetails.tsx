import { useContext } from 'react'
import { SelectedTimelineEntryContext } from '../../providers/selected-timeline-entry-provider'
import { formatTimestamp, formatTimestampShort } from '../../utils/date.utils'
import { TimelineEventDetailCall } from './TimelineEventDetailCall'

export function TimelineEventDetails() {
  const { selectedEntry } = useContext(SelectedTimelineEntryContext)


  if (!selectedEntry) {
    return (
      <div className='flex-1 p-4 overflow-auto flex items-center justify-center'>
        <span>No event selected</span>
      </div>
    )
  }

  return (
    <>
      <time
        dateTime={formatTimestampShort(selectedEntry.timeStamp)}
        className='flex-none py-0.5 text-xs leading-5 text-gray-500'
      >
        {formatTimestampShort(selectedEntry.timeStamp)}
      </time>
      <div className='pb-4'>
        Event type: {selectedEntry.entry?.case}
      </div>

      {(() => {
        switch (selectedEntry.entry?.case) {
          case 'call': return <TimelineEventDetailCall call={selectedEntry.entry.value} />
          default: return <></>
        }
      })()}
    </>

  )
}
