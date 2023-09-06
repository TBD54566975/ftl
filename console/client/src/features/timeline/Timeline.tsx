import {Timestamp} from '@bufbuild/protobuf'
import React from 'react'
import {useClient} from '../../hooks/use-client.ts'
import {ConsoleService} from '../../protos/xyz/block/ftl/v1/console/console_connect.ts'
import {
  LogEntry,
  StreamTimelineResponse,
} from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import {SidePanelContext} from '../../providers/side-panel-provider.tsx'
import {classNames} from '../../utils/react.utils.ts'
import {TimelineCall} from './TimelineCall.tsx'
import {TimelineDeployment} from './TimelineDeployment.tsx'
import {TimelineFilterBar} from './TimelineFilterBar.tsx'
import {TimelineLog} from './TimelineLog.tsx'
import {TimelineCallDetails} from './details/TimelineCallDetails.tsx'
import {TimelineDeploymentDetails} from './details/TimelineDeploymentDetails.tsx'
import {TimelineLogDetails} from './details/TimelineLogDetails.tsx'

/** A timeline entry with logs grouped together */
export class AnnotatedTimelineResponse extends StreamTimelineResponse {
  logs: LogEntry[] = []
}

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
    0, 1, 5, 9, 13, 17,
  ])

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

  const filteredEntries = entries
    .filter(entry => {
      const isActive = selectedEventTypes.includes(entry.entry?.case ?? '')
      if (entry.entry.case === 'log') {
        return (
          isActive && selectedLogLevels.includes(entry.entry.value.logLevel)
        )
      }

      return isActive
    })
    .map(entry => [
      {...entry, logs: []} as unknown as AnnotatedTimelineResponse,
    ])

  return (
    <div className='m-0'>
      <TimelineFilterBar
        selectedEventTypes={selectedEventTypes}
        onEventTypesChanged={handleEventTypesChanged}
        selectedLogLevels={selectedLogLevels}
        onLogLevelsChanged={handleLogLevelsChanged}
      />

      <ul
        role='list'
        className='space-y-4 p-4'
      >
        {filteredEntries
          // Group log entries together
          .reduce(
            (
              acc: AnnotatedTimelineResponse[],
              next: AnnotatedTimelineResponse[]
            ) => {
              if (acc.length === 0) return next
              const lastEntry = acc[acc.length - 1]
              const entry = next[0]
              const out = [...acc]
              if (
                lastEntry.entry?.case === 'log' &&
                lastEntry.logs.length === 0
              ) {
                lastEntry.logs = [lastEntry.entry.value]
              }
              if (
                lastEntry.entry?.case === 'log' &&
                entry.entry?.case === 'log'
              ) {
                lastEntry.logs.push(entry.entry.value)
              } else {
                out.push(entry)
              }
              return out
            },
            [] as AnnotatedTimelineResponse[]
          )
          .map((entry, index) => (
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
                        entry={entry}
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
