import type { Timestamp } from '@bufbuild/protobuf'
import { useContext, useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { timeFilter, useTimeline } from '../../api/timeline/index.ts'
import { Loader } from '../../components/Loader.tsx'
import type { Event, EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import { SidePanelContext } from '../../providers/side-panel-provider.tsx'
import { formatTimestampShort } from '../../utils/date.utils.ts'
import { panelColor } from '../../utils/style.utils.ts'
import { deploymentTextColor } from '../deployments/deployment.utils.ts'
import { TimelineCall } from './TimelineCall.tsx'
import { TimelineDeploymentCreated } from './TimelineDeploymentCreated.tsx'
import { TimelineDeploymentUpdated } from './TimelineDeploymentUpdated.tsx'
import { TimelineIcon } from './TimelineIcon.tsx'
import { TimelineLog } from './TimelineLog.tsx'
import { TimelineCallDetails } from './details/TimelineCallDetails.tsx'
import { TimelineDeploymentCreatedDetails } from './details/TimelineDeploymentCreatedDetails.tsx'
import { TimelineDeploymentUpdatedDetails } from './details/TimelineDeploymentUpdatedDetails.tsx'
import { TimelineLogDetails } from './details/TimelineLogDetails.tsx'
import type { TimeSettings } from './filters/TimelineTimeControls.tsx'

export const Timeline = ({ timeSettings, filters }: { timeSettings: TimeSettings; filters: EventsQuery_Filter[] }) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const { openPanel, closePanel, isOpen } = useContext(SidePanelContext)
  const [selectedEntry, setSelectedEntry] = useState<Event | null>(null)

  let eventFilters = filters
  if (timeSettings.newerThan || timeSettings.olderThan) {
    eventFilters = [timeFilter(timeSettings.olderThan, timeSettings.newerThan), ...filters]
  }

  const streamTimeline = timeSettings.isTailing && !timeSettings.isPaused

  const timeline = useTimeline(streamTimeline, eventFilters)

  useEffect(() => {
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
        openPanel(<TimelineCallDetails timestamp={entry.timeStamp as Timestamp} call={entry.entry.value} />, handlePanelClosed)
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

  const deploymentKey = (event: Event) => {
    switch (event.entry?.case) {
      case 'call':
        return event.entry.value.deploymentKey
      case 'log':
        return event.entry.value.deploymentKey
      case 'deploymentCreated':
        return event.entry.value.key
      case 'deploymentUpdated':
        return event.entry.value.key
      default:
        return ''
    }
  }

  if (timeline.isLoading) {
    return (
      <div className='flex justify-center items-center min-h-screen'>
        <Loader />
      </div>
    )
  }

  const entries = timeline.data || []

  return (
    <div className='border border-gray-100 dark:border-slate-700 rounded m-2'>
      <div className='overflow-x-hidden'>
        <table className={'w-full table-fixed text-gray-600 dark:text-gray-300'}>
          <thead>
            <tr className='flex text-xs'>
              <th className='p-1 text-left border-b w-8 border-gray-100 dark:border-slate-700 flex-none' />
              <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none'>Date</th>
              <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none'>Deployment</th>
              <th className='p-1 text-left border-b border-gray-100 dark:border-slate-700 flex-grow flex-shrink'>Content</th>
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
                <td className='p-1 w-40 items-center flex-none text-gray-400 dark:text-gray-400'>{formatTimestampShort(entry.timeStamp)}</td>
                <td title={deploymentKey(entry)} className={`p-1 pr-2 w-40 items-center flex-none truncate ${deploymentTextColor(deploymentKey(entry))}`}>
                  {deploymentKey(entry)}
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
