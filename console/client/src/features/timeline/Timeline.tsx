import { Timestamp } from '@bufbuild/protobuf'
import React from 'react'
import { Event, EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import { SidePanelContext } from '../../providers/side-panel-provider.tsx'
import { getEvents, streamEvents, timeFilter } from '../../services/console.service.ts'
import { formatTimestampShort } from '../../utils/date.utils.ts'
import { panelColor } from '../../utils/style.utils.ts'
import { TimelineCall } from './TimelineCall.tsx'
import { TimelineDeployment } from './TimelineDeployment.tsx'
import { TimelineIcon } from './TimelineIcon.tsx'
import { TimelineLog } from './TimelineLog.tsx'
import { TimelineCallDetails } from './details/TimelineCallDetails.tsx'
import { TimelineDeploymentDetails } from './details/TimelineDeploymentDetails.tsx'
import { TimelineLogDetails } from './details/TimelineLogDetails.tsx'
import { TimeSettings } from './filters/TimelineTimeControls.tsx'

interface Props {
  timeSettings: TimeSettings
  filters: EventsQuery_Filter[]
}

const maxTimelineEntries = 1000

export const Timeline = ({ timeSettings, filters }: Props) => {
  const { openPanel, closePanel, isOpen } = React.useContext(SidePanelContext)
  const [entries, setEntries] = React.useState<Event[]>([])
  const [selectedEntry, setSelectedEntry] = React.useState<Event | null>(null)
  const [selectedEventTypes] = React.useState<string[]>(['log', 'call', 'deployment'])
  const [selectedLogLevels] = React.useState<number[]>([1, 5, 9, 13, 17])

  React.useEffect(() => {
    const abortController = new AbortController()
    abortController.signal
    const fetchEvents = async () => {
      let eventFilters = filters
      if (timeSettings.newerThan || timeSettings.olderThan) {
        eventFilters = [timeFilter(timeSettings.olderThan, timeSettings.newerThan), ...filters]
      }
      const events = await getEvents(eventFilters)
      setEntries(events)
    }

    if (timeSettings.isTailing && !timeSettings.isPaused) {
      setEntries([])
      streamEvents({
        abortControllerSignal: abortController.signal,
        filters,
        onEventReceived: (event) => {
          if (!timeSettings.isPaused) {
            setEntries((prev) => [event, ...prev].slice(0, maxTimelineEntries))
          }
        },
      })
    } else {
      fetchEvents()
    }
    return () => {
      abortController.abort()
    }
  }, [filters, timeSettings])

  React.useEffect(() => {
    if (!isOpen) {
      setSelectedEntry(null)
    }
  }, [isOpen])

  const handleEntryClicked = (entry: Event) => {
    if (selectedEntry === entry) {
      setSelectedEntry(null)
      closePanel()
      return
    }

    switch (entry.entry?.case) {
      case 'call':
        openPanel(<TimelineCallDetails timestamp={entry.timeStamp as Timestamp} call={entry.entry.value} />)
        break
      case 'log':
        openPanel(<TimelineLogDetails event={entry} log={entry.entry.value} />)
        break
      case 'deployment':
        openPanel(<TimelineDeploymentDetails event={entry} deployment={entry.entry.value} />)
        break
      default:
        break
    }
    setSelectedEntry(entry)
  }

  const filteredEntries = entries.filter((entry) => {
    const isActive = selectedEventTypes.includes(entry.entry?.case ?? '')
    if (entry.entry.case === 'log') {
      return isActive && selectedLogLevels.includes(entry.entry.value.logLevel)
    }

    return isActive
  })

  return (
    <div className='border border-gray-100 dark:border-slate-700 rounded m-2'>
      <div className='overflow-x-hidden'>
        <table className={`w-full table-fixed text-gray-600 dark:text-gray-300`}>
          <thead>
            <tr className='flex text-xs'>
              <th className='p-1 text-left border-b w-8 border-gray-100 dark:border-slate-700 flex-none'></th>
              <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none'>Date</th>
              <th className='p-1 text-left border-b border-gray-100 dark:border-slate-700 flex-grow flex-shrink'>
                Content
              </th>
            </tr>
          </thead>
          <tbody>
            {filteredEntries.map((entry) => (
              <tr
                key={entry.id.toString()}
                className={`flex border-b border-gray-100 dark:border-slate-700 text-xs font-roboto-mono ${
                  selectedEntry?.id === entry.id ? 'bg-indigo-50 dark:bg-slate-700' : panelColor
                } relative flex cursor-pointer hover:bg-indigo-50 dark:hover:bg-slate-700`}
                onClick={() => handleEntryClicked(entry)}
              >
                <td className='w-8 flex-none flex items-center justify-center'>
                  <TimelineIcon event={entry} />
                </td>
                <td className='p-1 w-40 items-center flex-none text-gray-400 dark:text-gray-400'>
                  {formatTimestampShort(entry.timeStamp)}
                </td>
                <td className='p-1 flex-grow truncate'>
                  {(() => {
                    switch (entry.entry?.case) {
                      case 'call':
                        return <TimelineCall call={entry.entry.value} />
                      case 'log':
                        return <TimelineLog log={entry.entry.value} />
                      case 'deployment':
                        return <TimelineDeployment deployment={entry.entry.value} />
                    }
                  })()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
