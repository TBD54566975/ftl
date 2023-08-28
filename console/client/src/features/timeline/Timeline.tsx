import {Timestamp} from '@bufbuild/protobuf'
import React from 'react'
import {useClient} from '../../hooks/use-client.ts'
import {ConsoleService} from '../../protos/xyz/block/ftl/v1/console/console_connect.ts'
import {StreamTimelineResponse} from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import {SidePanelContext} from '../../providers/side-panel-provider.tsx'
import {classNames} from '../../utils/react.utils.ts'
import {TimelineCall} from './TimelineCall.tsx'
import {TimelineDeployment} from './TimelineDeployment.tsx'
import {TimelineFilterBar} from './TimelineFilterBar.tsx'
import {TimelineLog} from './TimelineLog.tsx'
import {TimelineCallDetails} from './details/TimelineCallDetails.tsx'
import {TimelineDeploymentDetails} from './details/TimelineDeploymentDetails.tsx'
import {TimelineLogDetails} from './details/TimelineLogDetails.tsx'

export const Timeline = () => {
  const client = useClient(ConsoleService)
  const {openPanel, closePanel, isOpen} = React.useContext(SidePanelContext)
  const [entries, setEntries] = React.useState<StreamTimelineResponse[]>([])
  const [selectedEntry, setSelectedEntry] =
    React.useState<StreamTimelineResponse | null>(null)
  const [selectedFilters, setSelectedFilters] = React.useState<string[]>([])

  React.useEffect(() => {
    const abortController = new AbortController()

    async function streamTimeline() {
      const afterTime = new Date()
      afterTime.setHours(afterTime.getHours() - 1)

      for await (const response of client.streamTimeline(
        {afterTime: Timestamp.fromDate(afterTime)},
        {signal: abortController.signal}
      )) {
        if (response.entry) {
          setEntries(prevEntries => [response, ...prevEntries])
        }
      }
    }

    void streamTimeline()
    return () => {
      abortController.abort()
    }
  }, [client])

  React.useEffect(() => {
    if (!isOpen) {
      setSelectedEntry(null)
    }
  }, [isOpen])

  const handleEntryClicked = (entry: StreamTimelineResponse) => {
    if (selectedEntry === entry) {
      setSelectedEntry(null)
      closePanel()
      return
    }

    switch (entry.entry?.case) {
      case 'call':
        openPanel(
          <TimelineCallDetails
            timestamp={entry.timeStamp as Timestamp}
            call={entry.entry.value}
          />
        )
        break
      case 'log':
        openPanel(
          <TimelineLogDetails
            entry={entry}
            log={entry.entry.value}
          />
        )
        break
      case 'deployment':
        openPanel(
          <TimelineDeploymentDetails
            entry={entry}
            deployment={entry.entry.value}
          />
        )
        break
      default:
        break
    }
    setSelectedEntry(entry)
  }

  const handleFilterChange = (optionValue: string, checked: boolean) => {
    if (checked) {
      setSelectedFilters(prev => [...prev, optionValue])
    } else {
      setSelectedFilters(prev => prev.filter(filter => filter !== optionValue))
    }
  }

  const filteredEntries = entries.filter(entry => {
    if (selectedFilters.length === 0) {
      return true
    }
    return selectedFilters.includes(entry.entry?.case ?? '')
  })

  return (
    <div className='m-0'>
      <TimelineFilterBar
        selectedFilters={selectedFilters}
        onFilterChange={handleFilterChange}
      />

      <ul
        role='list'
        className='space-y-4 p-4'
      >
        {filteredEntries.map((entry, index) => (
          <li
            key={entry.id.toString()}
            className='relative flex gap-x-4'
            onClick={() => handleEntryClicked(entry)}
          >
            <div
              className={classNames(
                index === filteredEntries.length - 1 ? 'h-6' : '-bottom-6',
                'absolute left-0 top-0 flex w-6 justify-center'
              )}
            >
              <div className='w-px bg-gray-200 dark:bg-gray-600' />
            </div>

            {(() => {
              switch (entry.entry?.case) {
                case 'call':
                  return (
                    <TimelineCall
                      call={entry.entry.value}
                      selected={selectedEntry === entry}
                    />
                  )
                case 'log':
                  return (
                    <TimelineLog
                      log={entry.entry.value}
                      selected={selectedEntry === entry}
                    />
                  )
                case 'deployment':
                  return (
                    <TimelineDeployment
                      deployment={entry.entry.value}
                      timestamp={entry.timeStamp}
                      selected={selectedEntry === entry}
                    />
                  )
                default:
                  return <></>
              }
            })()}
          </li>
        ))}
      </ul>
    </div>
  )
}
