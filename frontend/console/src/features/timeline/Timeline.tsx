import { useContext, useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { timeFilter, useTimeline } from '../../api/timeline/index.ts'
import { Loader } from '../../components/Loader.tsx'
import type { Event } from '../../protos/xyz/block/ftl/timeline/v1/event_pb.ts'
import type { GetTimelineRequest_Filter } from '../../protos/xyz/block/ftl/timeline/v1/timeline_pb.ts'
import { SidePanelContext } from '../../providers/side-panel-provider.tsx'
import TimelineEventList from './TimelineEventList.tsx'
import { TimelineAsyncExecuteDetails } from './details/TimelineAsyncExecuteDetails.tsx'
import { TimelineCallDetails } from './details/TimelineCallDetails.tsx'
import { TimelineCronScheduledDetails } from './details/TimelineCronScheduledDetails.tsx'
import { TimelineDeploymentCreatedDetails } from './details/TimelineDeploymentCreatedDetails.tsx'
import { TimelineDeploymentUpdatedDetails } from './details/TimelineDeploymentUpdatedDetails.tsx'
import { TimelineDetailsHeader } from './details/TimelineDetailsHeader.tsx'
import { TimelineIngressDetails } from './details/TimelineIngressDetails.tsx'
import { TimelineLogDetails } from './details/TimelineLogDetails.tsx'
import { TimelinePubSubConsumeDetails } from './details/TimelinePubSubConsumeDetails.tsx'
import { TimelinePubSubPublishDetails } from './details/TimelinePubSubPublishDetails.tsx'
import type { TimeSettings } from './filters/TimelineTimeControls.tsx'

export const Timeline = ({ timeSettings, filters }: { timeSettings: TimeSettings; filters: GetTimelineRequest_Filter[] }) => {
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
        openPanel(<TimelineCallDetails event={entry} />, <TimelineDetailsHeader event={entry} />, handlePanelClosed)
        break
      case 'log':
        openPanel(<TimelineLogDetails event={entry} log={entry.entry.value} />, <TimelineDetailsHeader event={entry} />, handlePanelClosed)
        break
      case 'deploymentCreated':
        openPanel(<TimelineDeploymentCreatedDetails event={entry} deployment={entry.entry.value} />, <TimelineDetailsHeader event={entry} />, handlePanelClosed)
        break
      case 'deploymentUpdated':
        openPanel(<TimelineDeploymentUpdatedDetails event={entry} deployment={entry.entry.value} />, <TimelineDetailsHeader event={entry} />, handlePanelClosed)
        break
      case 'ingress':
        openPanel(<TimelineIngressDetails event={entry} />, <TimelineDetailsHeader event={entry} />, handlePanelClosed)
        break
      case 'cronScheduled':
        openPanel(<TimelineCronScheduledDetails event={entry} />, <TimelineDetailsHeader event={entry} />, handlePanelClosed)
        break
      case 'asyncExecute':
        openPanel(<TimelineAsyncExecuteDetails event={entry} />, <TimelineDetailsHeader event={entry} />, handlePanelClosed)
        break
      case 'pubsubPublish':
        openPanel(<TimelinePubSubPublishDetails event={entry} />, <TimelineDetailsHeader event={entry} />, handlePanelClosed)
        break
      case 'pubsubConsume':
        openPanel(<TimelinePubSubConsumeDetails event={entry} />, <TimelineDetailsHeader event={entry} />, handlePanelClosed)
        break
      default:
        break
    }
    setSelectedEntry(entry)
    setSearchParams({ ...Object.fromEntries(searchParams.entries()), id: entry.id.toString() })
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
      <TimelineEventList events={entries} selectedEventId={selectedEntry?.id} handleEntryClicked={handleEntryClicked} />
    </div>
  )
}
