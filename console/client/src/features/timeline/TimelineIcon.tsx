import {
  Call,
  ListAlt,
  ListOutlined,
  PhoneCallback,
  RocketLaunch,
} from '@mui/icons-material'
import {StreamTimelineResponse} from '../../protos/xyz/block/ftl/v1/console/console_pb'

type Props = {
  entry: StreamTimelineResponse
}

export const logLevelIconColor = {
  1: 'text-indigo-600 dark:text-indigo-600',
  5: 'text-indigo-600 dark:text-indigo-600',
  9: 'text-green-500 dark:text-green-400',
  13: 'text-yellow-400 dark:text-yellow-300',
  17: 'text-red-500 dark:text-red-400',
}

export const TimelineIcon: React.FC<Props> = ({entry}) => {
  const iconColor = (entry: StreamTimelineResponse) => {
    switch (entry.entry.case) {
      case 'call':
        return entry.entry.value.error ? 'text-red-600' : 'text-indigo-600'
      case 'log':
        return `${logLevelIconColor[entry.entry.value.logLevel]}`
      default:
        return 'text-indigo-600'
    }
  }

  const icon = (entry: StreamTimelineResponse) => {
    const iconSize = 20
    switch (entry.entry.case) {
      case 'call':
        return entry.entry.value.sourceVerbRef ? (
          <PhoneCallback sx={{fontSize: iconSize}} />
        ) : (
          <Call sx={{fontSize: iconSize}} />
        )
      case 'deployment':
        return <RocketLaunch sx={{fontSize: iconSize}} />
      case 'log':
        return <ListOutlined sx={{fontSize: iconSize}} />
      default:
        return <ListAlt sx={{fontSize: iconSize}} />
    }
  }

  return <div className={`${iconColor(entry)}`}>{icon(entry)}</div>
}
