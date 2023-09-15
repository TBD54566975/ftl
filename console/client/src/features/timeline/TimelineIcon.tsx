import { ListBulletIcon, PhoneArrowDownLeftIcon, PhoneIcon, RocketLaunchIcon } from '@heroicons/react/24/outline'
import { TimelineEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { LogLevelBadgeSmall } from '../logs/LogLevelBadgeSmall'

interface Props {
  entry: TimelineEvent
}

export const TimelineIcon = ({ entry }: Props) => {
  const icon = (entry: TimelineEvent) => {
    const style = 'h4 w-4 text-indigo-600'
    switch (entry.entry.case) {
      case 'call': {
        const textColor = entry.entry.value.error ? 'text-red-600' : 'text-indigo-600'
        return entry.entry.value.sourceVerbRef ? (
          <PhoneIcon className={`${style} ${textColor}`} />
        ) : (
          <PhoneArrowDownLeftIcon className={`${style}`} />
        )
      }
      case 'deployment':
        return <RocketLaunchIcon className={`${style}`} />
      case 'log':
        return <LogLevelBadgeSmall logLevel={entry.entry.value.logLevel} />
      default:
        return <ListBulletIcon className={`${style}`} />
    }
  }

  return <div>{icon(entry)}</div>
}
