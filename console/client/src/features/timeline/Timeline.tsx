import { Timestamp } from '@bufbuild/protobuf'
import React from 'react'
import { useSearchParams } from 'react-router-dom'
import { Event, EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import { SidePanelContext } from '../../providers/side-panel-provider.tsx'
import { eventIdFilter, getEvents, streamEvents, timeFilter } from '../../services/console.service.ts'
import { formatTimestampShort } from '../../utils/date.utils.ts'
import { panelColor } from '../../utils/style.utils.ts'
import { TimelineCall } from './TimelineCall.tsx'
import { TimelineDeploymentCreated } from './TimelineDeploymentCreated.tsx'
import { TimelineDeploymentUpdated } from './TimelineDeploymentUpdated.tsx'
import { TimelineIcon } from './TimelineIcon.tsx'
import { TimelineLog } from './TimelineLog.tsx'
import { TimelineCallDetails } from './details/TimelineCallDetails.tsx'
import { TimelineDeploymentCreatedDetails } from './details/TimelineDeploymentCreatedDetails.tsx'
import { TimelineDeploymentUpdatedDetails } from './details/TimelineDeploymentUpdatedDetails.tsx'
import { TimelineLogDetails } from './details/TimelineLogDetails.tsx'
import { TimeSettings } from './filters/TimelineTimeControls.tsx'

interface Props {
  timeSettings: TimeSettings
  filters: EventsQuery_Filter[]
}

const maxTimelineEntries = 1000

export const Timeline = ({ timeSettings, filters }: Props) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const { openPanel, closePanel, isOpen } = React.useContext(SidePanelContext)
  const [entries, setEntries] = React.useState<Event[]>([])
  const [selectedEntry, setSelectedEntry] = React.useState<Event | null>(null)

  React.useEffect(() => {
    const eventId = searchParams.get('id')
    const abortController = new AbortController()

    const fetchEvents = async () => {
      let eventFilters = filters
      if (timeSettings.newerThan || timeSettings.olderThan) {
        eventFilters = [timeFilter(timeSettings.olderThan, timeSettings.newerThan), ...filters]
      }

      if (eventId) {
        const id = BigInt(eventId)
        eventFilters = [eventIdFilter({ higherThan: id }), ...filters]
      }
      const events = await getEvents({ abortControllerSignal: abortController.signal, filters: eventFilters })
      setEntries(events)

      if (eventId) {
        const entry = events.find((event) => event.id.toString() === eventId)
        if (entry) {
          handleEntryClicked(entry)
        }
      }
    }

    if (timeSettings.isTailing && !timeSettings.isPaused && !eventId) {
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

  const handlePanelClosed = () => {
    const newParams = new URLSearchParams(searchParams.toString())
    newParams.delete('id')
    setSearchParams(newParams)
    setSelectedEntry(null)
  }

  const handleEntryClicked = (entry: Event) => {
    if (selectedEntry === entry) {
      closePanel()
      return
    }

    switch (entry.entry?.case) {
      case 'call':
        openPanel(
          <TimelineCallDetails timestamp={entry.timeStamp as Timestamp} call={entry.entry.value} />,
          handlePanelClosed,
        )
        break
      case 'log':
        openPanel(<TimelineLogDetails event={entry} log={entry.entry.value} />, handlePanelClosed)
        break
      case 'deploymentCreated':
        openPanel(<TimelineDeploymentCreatedDetails event={entry} deployment={entry.entry.value} />, handlePanelClosed)
        break
      case 'deploymentUpdated':
        openPanel(<TimelineDeploymentUpdatedDetails event={entry} deployment={entry.entry.value} />, handlePanelClosed)
        break
      default:
        break
    }
    setSelectedEntry(entry)
    setSearchParams({ ...Object.fromEntries(searchParams.entries()), id: entry.id.toString() })
  }

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
            {entries.map((entry) => (
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
                      case 'deploymentCreated':
                        return <TimelineDeploymentCreated deployment={entry.entry.value} />
                      case 'deploymentUpdated':
                        return <TimelineDeploymentUpdated deployment={entry.entry.value} />
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
