import { useContext } from 'react'
import { SelectedTimelineEntryContext } from '../../providers/selected-timeline-entry-provider'
import { formatTimestampShort } from '../../utils/date.utils'
import { TimelineEventDetailCall } from './TimelineEventDetailCall'
import { TimelineEventDetailLog } from './TimelineEventDetailLog'
import { lightTextColor, textColor } from '../../utils/style.utils'

export function TimelineEventDetails() {
  const { selectedEntry } = useContext(SelectedTimelineEntryContext)


  if (!selectedEntry) {
    return (
      <> </>
    )
  }

  return (
    <>
      <time
        dateTime={formatTimestampShort(selectedEntry.timeStamp)}
        className={`flex-none py-0.5 text-xs leading-5 ${lightTextColor}`}
      >
        {formatTimestampShort(selectedEntry.timeStamp)}
      </time>
      <div className={`py-2 ${textColor}`}>
        Event type: {selectedEntry.entry?.case}
      </div>

      {(() => {
        switch (selectedEntry.entry?.case) {
          case 'call': return <TimelineEventDetailCall call={selectedEntry.entry.value} />
          case 'log': return <TimelineEventDetailLog log={selectedEntry.entry.value} />
          default: return <></>
        }
      })()}
    </>

  )
}
