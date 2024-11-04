import { useContext } from 'react'
import { CloseButton } from '../../../components'
import { Divider } from '../../../components/Divider'
import type { Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../../providers/side-panel-provider'
import { formatTimestampShort } from '../../../utils'
import { logLevelBadge, logLevelText } from '../../logs/log.utils'
import { refString } from '../../verbs/verb.utils'
import { TimelineDetailsColorBar } from './TimelineDetailsColorBar'

export const TimelineDetailsHeader = ({ event }: { event: Event }) => {
  const { closePanel } = useContext(SidePanelContext)

  return (
    <div>
      <TimelineDetailsColorBar event={event} />
      <div className='p-4'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center space-x-2'>
            {eventBadge(event)}

            <time dateTime={formatTimestampShort(event.timeStamp)} className='flex-none text-sm font-roboto-mono text-gray-500 dark:text-gray-300'>
              {formatTimestampShort(event.timeStamp)}
            </time>
          </div>
          <CloseButton onClick={closePanel} />
        </div>
      </div>
      <Divider />
    </div>
  )
}

const eventBadge = (event: Event) => {
  switch (event.entry?.case) {
    case 'call':
      return (
        <div className={'inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100'}>
          {refString(event.entry.value.destinationVerbRef)}
        </div>
      )
    case 'asyncExecute':
      return (
        <div className={'inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100'}>
          {refString(event.entry.value.verbRef)}
        </div>
      )
    case 'cronScheduled':
      return (
        <div className={'inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100'}>
          {refString(event.entry.value.verbRef)}
        </div>
      )
    case 'log':
      return (
        <span className={`${logLevelBadge[event.entry.value.logLevel]} inline-flex items-center rounded-md px-2 py-1 text-xs font-medium font-roboto-mono`}>
          {logLevelText[event.entry.value.logLevel]}
        </span>
      )
    case 'deploymentCreated':
      return (
        <span
          className={
            'text-green-500 bg-green-400/30 dark:text-green-300 dark:bg-green-700/10 inline-flex items-center rounded-md px-2 py-1 text-xs font-medium'
          }
        >
          Deployment Created
        </span>
      )
    case 'deploymentUpdated':
      return (
        <span
          className={'text-blue-500 bg-blue-300/30 dark:text-blue-300 dark:bg-blue-700/30 inline-flex items-center rounded-md px-2 py-1 text-xs font-medium'}
        >
          Deployment Updated
        </span>
      )
    case 'ingress':
      return (
        <div className={'inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100'}>
          {event.entry.value.path}
        </div>
      )
    case 'pubsubPublish':
      return (
        <div className={'inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100'}>
          {event.entry.value.topic}
        </div>
      )
    case 'pubsubConsume':
      return (
        <div className={'inline-block rounded-md bg-indigo-200 dark:bg-indigo-700 px-2 py-1 mr-1 text-sm font-medium text-gray-700 dark:text-gray-100'}>
          {event.entry.value.topic}
        </div>
      )
    default:
      return ''
  }
}
