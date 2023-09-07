import {Timestamp} from '@bufbuild/protobuf'
import React from 'react'
import {useClient} from '../../hooks/use-client.ts'
import {ConsoleService} from '../../protos/xyz/block/ftl/v1/console/console_connect.ts'
import {StreamTimelineResponse} from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import {SidePanelContext} from '../../providers/side-panel-provider.tsx'
import {classNames} from '../../utils/react.utils.ts'
import {TimelineCall} from './TimelineCall.tsx'
import {TimelineDeployment} from './TimelineDeployment.tsx'
import {TimelineLog} from './TimelineLog.tsx'
import {TimelineCallDetails} from './details/TimelineCallDetails.tsx'
import {TimelineDeploymentDetails} from './details/TimelineDeploymentDetails.tsx'
import {TimelineLogDetails} from './details/TimelineLogDetails.tsx'
import {TIME_RANGES} from './filters/TimeFilter.tsx'
import {TimelineFilterBar} from './filters/TimelineFilterBar.tsx'

export const Timeline = () => {
  const client = useClient(ConsoleService)
  const {openPanel, closePanel, isOpen} = React.useContext(SidePanelContext)
  const [entries, setEntries] = React.useState<StreamTimelineResponse[]>([])
  const [selectedEntry, setSelectedEntry] =
    React.useState<StreamTimelineResponse | null>(null)
  const [selectedEventTypes, setSelectedEventTypes] = React.useState<string[]>([
    'log',
    'call',
    'deployment',
  ])
  const [selectedLogLevels, setSelectedLogLevels] = React.useState<number[]>([
    1, 5, 9, 13, 17,
  ])
  const [selectedTimeRange, setSelectedTimeRange] = React.useState('1h')

  React.useEffect(() => {
    const abortController = new AbortController()

    async function streamTimeline() {
      setEntries(_ => [])
      const afterTime = new Date(
        Date.now() - TIME_RANGES[selectedTimeRange].value
      )

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
  }, [client, selectedTimeRange])

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

  const handleEventTypesChanged = (eventType: string, checked: boolean) => {
    if (checked) {
      setSelectedEventTypes(prev => [...prev, eventType])
    } else {
      setSelectedEventTypes(prev => prev.filter(filter => filter !== eventType))
    }
  }

  const handleLogLevelsChanged = (logLevel: number, checked: boolean) => {
    if (checked) {
      setSelectedLogLevels(prev => [...prev, logLevel])
    } else {
      setSelectedLogLevels(prev => prev.filter(filter => filter !== logLevel))
    }
  }

  const handleTimeRangeChanged = (key: string) => {
    setSelectedTimeRange(key)
  }

  const filteredEntries = entries.filter(entry => {
    const isActive = selectedEventTypes.includes(entry.entry?.case ?? '')
    if (entry.entry.case === 'log') {
      return isActive && selectedLogLevels.includes(entry.entry.value.logLevel)
    }

    return isActive
  })

  return (
    <div className='m-0'>
      <TimelineFilterBar
        selectedEventTypes={selectedEventTypes}
        onEventTypesChanged={handleEventTypesChanged}
        selectedLogLevels={selectedLogLevels}
        onLogLevelsChanged={handleLogLevelsChanged}
        selectedTimeRange={selectedTimeRange}
        onSelectedTimeRangeChanged={handleTimeRangeChanged}
      />

      <ul
        role='list'
        className='space-y-4 p-4'
      >
        {filteredEntries.map((entry, index) => (
          <li
            key={entry.id.toString()}
            className='relative flex gap-x-4 cursor-pointer'
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
