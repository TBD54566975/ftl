import { useContext } from 'react'
import { SelectedTimelineEntryContext } from '../../providers/selected-timeline-entry-provider'
import { formatTimestampShort } from '../../utils/date.utils'

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
    <div className='text-sm'>
      Event timestamp: {formatTimestampShort(selectedEntry.timeStamp)}
    </div>
  )
}
