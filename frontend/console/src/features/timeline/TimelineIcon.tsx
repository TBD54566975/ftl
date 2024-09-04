import { Call02Icon, CallIncoming04Icon, Menu01Icon, Rocket01Icon } from 'hugeicons-react'
import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { LogLevelBadgeSmall } from '../logs/LogLevelBadgeSmall'

export const TimelineIcon = ({ event }: { event: Event }) => {
  const icon = (event: Event) => {
    const style = 'h4 w-4 text-indigo-600'
    switch (event.entry.case) {
      case 'call': {
        const textColor = event.entry.value.error ? 'text-red-600' : 'text-indigo-600'
        return event.entry.value.sourceVerbRef ? <Call02Icon className={`${style} ${textColor}`} /> : <CallIncoming04Icon className={`${style} ${textColor}`} />
      }
      case 'deploymentCreated':
        return <Rocket01Icon className='h4 w-4 text-green-500' />
      case 'deploymentUpdated':
        return <Rocket01Icon className='h4 w-4 text-indigo-600' />
      case 'log':
        return <LogLevelBadgeSmall logLevel={event.entry.value.logLevel} />
      default:
        return <Menu01Icon className={`${style}`} />
    }
  }

  return <div>{icon(event)}</div>
}
