import { ListBulletIcon, PhoneArrowDownLeftIcon, PhoneIcon, RocketLaunchIcon } from '@heroicons/react/24/outline'
import { TimelineEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'

interface Props {
  entry: TimelineEvent
}

export const logLevelIconColor: { [key: number]: string } = {
  1: 'text-indigo-600 dark:text-indigo-600',
  5: 'text-indigo-600 dark:text-indigo-600',
  9: 'text-green-500 dark:text-green-400',
  13: 'text-yellow-400 dark:text-yellow-300',
  17: 'text-red-500 dark:text-red-400',
}

export const TimelineIcon = ({ entry }: Props) => {
  const iconColor = (entry: TimelineEvent) => {
    switch (entry.entry.case) {
      case 'call':
        return entry.entry.value.error ? 'text-red-600' : 'text-indigo-600'
      case 'log':
        return `${logLevelIconColor[entry.entry.value.logLevel]}`
      default:
        return 'text-indigo-600'
    }
  }

  const icon = (entry: TimelineEvent) => {
    const style = 'h4 w-4'
    switch (entry.entry.case) {
      case 'call':
        return entry.entry.value.sourceVerbRef ? (
          <PhoneIcon className={`${style}`} />
        ) : (
          <PhoneArrowDownLeftIcon className={`${style}`} />
        )
      case 'deployment':
        return <RocketLaunchIcon className={`${style}`} />
      case 'log':
        return <ListBulletIcon className={`${style}`} />
      default:
        return <ListBulletIcon className={`${style}`} />
    }
  }

  return <div className={`${iconColor(entry)}`}>{icon(entry)}</div>
}
