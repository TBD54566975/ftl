import React from 'react'
import { useSearchParams } from 'react-router-dom'
import { Panel } from './components'
import { modulesFilter } from '../../services/console.service'
import { streamEvents } from '../../services/console.service'
import { Event, Module, EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { TimelineCall } from '../timeline/TimelineCall.tsx'
import { TimelineDeploymentCreated } from '../timeline/TimelineDeploymentCreated.tsx'
import { TimelineDeploymentUpdated } from '../timeline/TimelineDeploymentUpdated.tsx'
import { TimelineIcon } from '../timeline/TimelineIcon.tsx'
import { TimelineLog } from '../timeline/TimelineLog.tsx'
import { TimelineCallDetails } from '../timeline/details/TimelineCallDetails.tsx'
import { TimelineDeploymentCreatedDetails } from '../timeline/details/TimelineDeploymentCreatedDetails.tsx'
import { TimelineDeploymentUpdatedDetails } from '../timeline/details/TimelineDeploymentUpdatedDetails.tsx'
import { TimelineLogDetails } from '../timeline/details/TimelineLogDetails.tsx'
import { SidePanelContext } from '../../providers/side-panel-provider.tsx'
import { formatTimestampShort } from '../../utils/date.utils.ts'
import { panelColor } from '../../utils/style.utils.ts'

const maxTimelineEntries = 1000

export const ModulesRequests: React.FC<{
  className?: string
  modules: Module[]
}> = ({ className, modules }) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const [entries, setEntries] = React.useState<Event[]>([])
  const { openPanel, closePanel } = React.useContext(SidePanelContext)
  const [selectedEntry, setSelectedEntry] = React.useState<Event | null>(null)
  const deployments = modules.map(({ deploymentName }) => deploymentName)

  const filters: EventsQuery_Filter[] = []
  if (deployments.length) {
    filters.push(modulesFilter(deployments))
  }

  React.useEffect(() => {
    setEntries([])
    if (!filters.length) return
    const abortController = new AbortController()
    abortController.signal
    streamEvents({
      abortControllerSignal: abortController.signal,
      filters,
      onEventReceived: (event) => {
        setEntries((prev) => [event, ...prev].slice(0, maxTimelineEntries))
      },
    })
    return () => {
      abortController.abort()
    }
  }, [modules])
  const handleEntryClicked = (entry: Event) => {
    if (selectedEntry === entry) {
      setSelectedEntry(null)
      closePanel()
      const newParams = new URLSearchParams(searchParams.toString())
      newParams.delete('id')
      setSearchParams(newParams)
      return
    }

    switch (entry.entry?.case) {
      case 'call':
        openPanel(<TimelineCallDetails timestamp={entry.timeStamp as Timestamp} call={entry.entry.value} />)
        break
      case 'log':
        openPanel(<TimelineLogDetails event={entry} log={entry.entry.value} />)
        break
      case 'deploymentCreated':
        openPanel(<TimelineDeploymentCreatedDetails event={entry} deployment={entry.entry.value} />)
        break
      case 'deploymentUpdated':
        openPanel(<TimelineDeploymentUpdatedDetails event={entry} deployment={entry.entry.value} />)
        break
      default:
        break
    }
    setSelectedEntry(entry)
    setSearchParams({ ...Object.fromEntries(searchParams.entries()), id: entry.id.toString() })
  }
  return (
    <Panel className={className}>
      <Panel.Header className='flex border-gray-100 dark:border-slate-700 '>
        <span className='p-1 text-left border-b w-8 flex-none'></span>
        <span className='p-1 text-left border-b w-40 flex-none'>Date</span>
        <span className='p-1 text-left border-b flex-grow flex-shrink'>Content</span>
      </Panel.Header>
      <Panel.Body>
        {entries.map((entry) => (
          <div
            key={entry.id.toString()}
            className={`flex border-b border-gray-100 dark:border-slate-700 text-xs font-roboto-mono ${
              selectedEntry?.id === entry.id ? 'bg-indigo-50 dark:bg-slate-700' : panelColor
            } relative flex cursor-pointer hover:bg-indigo-50 dark:hover:bg-slate-700`}
            onClick={() => handleEntryClicked(entry)}
          >
            <span className='w-8 flex-none flex items-center justify-center'>
              <TimelineIcon event={entry} />
            </span>
            <span className='p-1 w-40 items-center flex-none text-gray-400 dark:text-gray-400'>
              {formatTimestampShort(entry.timeStamp)}
            </span>
            <span className='p-1 flex-grow truncate'>
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
            </span>
          </div>
        ))}
      </Panel.Body>
    </Panel>
  )
}
