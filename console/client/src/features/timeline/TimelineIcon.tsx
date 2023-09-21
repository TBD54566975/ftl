import { ListBulletIcon, PhoneArrowDownLeftIcon, PhoneIcon, RocketLaunchIcon } from '@heroicons/react/24/outline'
import { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { LogLevelBadgeSmall } from '../logs/LogLevelBadgeSmall'

interface Props {
  event: Event
}

export const TimelineIcon = ({ event }: Props) => {
  const icon = (event: Event) => {
    const style = 'h4 w-4 text-indigo-600'
    switch (event.entry.case) {
      case 'call': {
        const textColor = event.entry.value.error ? 'text-red-600' : 'text-indigo-600'
        return event.entry.value.sourceVerbRef ? (
          <PhoneIcon className={`${style} ${textColor}`} />
        ) : (
          <PhoneArrowDownLeftIcon className={`${style}`} />
        )
      }
      case 'deploymentCreated':
        return <RocketLaunchIcon className='h4 w-4 text-green-500' />
      case 'deploymentUpdated':
        return <RocketLaunchIcon className='h4 w-4 text-indigo-600' />
      case 'log':
        return <LogLevelBadgeSmall logLevel={event.entry.value.logLevel} />
      default:
        return <ListBulletIcon className={`${style}`} />
    }
  }

  return <div>{icon(event)}</div>
}
