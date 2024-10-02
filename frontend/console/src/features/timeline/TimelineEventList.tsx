import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatTimestampShort } from '../../utils'
import { deploymentTextColor } from '../deployments/deployment.utils'
import { TimelineCall } from './TimelineCall'
import { TimelineDeploymentCreated } from './TimelineDeploymentCreated'
import { TimelineDeploymentUpdated } from './TimelineDeploymentUpdated'
import { TimelineIcon } from './TimelineIcon'
import { TimelineIngress } from './TimelineIngress'
import { TimelineLog } from './TimelineLog'

interface EventTimelineProps {
  events: Event[]
  selectedEventId?: bigint
  handleEntryClicked: (entry: Event) => void
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
    case 'ingress':
      return event.entry.value.deploymentKey
    default:
      return ''
  }
}

export const TimelineEventList = ({ events, selectedEventId, handleEntryClicked }: EventTimelineProps) => {
  return (
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
          {events.map((entry) => (
            <tr
              key={entry.id}
              className={`flex border-b border-gray-100 dark:border-slate-700 text-xs font-roboto-mono ${
                selectedEventId === entry.id ? 'bg-indigo-50 dark:bg-slate-700' : ''
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
                  switch (entry.entry.case) {
                    case 'call':
                      return <TimelineCall call={entry.entry.value} />
                    case 'log':
                      return <TimelineLog log={entry.entry.value} />
                    case 'deploymentCreated':
                      return <TimelineDeploymentCreated deployment={entry.entry.value} />
                    case 'deploymentUpdated':
                      return <TimelineDeploymentUpdated deployment={entry.entry.value} />
                    case 'ingress':
                      return <TimelineIngress ingress={entry.entry.value} />
                    default:
                      return null
                  }
                })()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default TimelineEventList
