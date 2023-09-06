import {formatTimestamp} from '../../utils/date.utils'
import {classNames} from '../../utils/react.utils'
import {
  logLevelBadge,
  logLevelText,
  panelColor,
  textColor,
} from '../../utils/style.utils'
import {DocumentTextIcon} from '@heroicons/react/24/outline'
import {AnnotatedTimelineResponse} from './Timeline.tsx'

type Props = {
  entry: AnnotatedTimelineResponse
  selected?: boolean
}

export const TimelineLog: React.FC<Props> = ({entry, selected}) => {
  return (
    <>
      <div
        className={`relative flex w-6 flex-none items-top justify-center ${panelColor}`}
      >
        <DocumentTextIcon
          className={`h-6 w-6 text-indigo-500`}
          aria-hidden='true'
        />
      </div>
      <ul
        role='list'
        className='space-y-1'
      >
        {entry.logs.map(log => (
          <li
            key={entry.id.toString()}
            className='relative flex gap-x-2'
          >
            <span
              className={classNames(
                `${logLevelBadge[log.logLevel]}`,
                'inline-flex items-center justify-center rounded-md px-2 text-xs font-medium text-gray-600 w-12'
              )}
            >
              {logLevelText[log.logLevel]}
            </span>
            <pre
              className={`flex-auto text-xs leading-5 ${
                selected ? 'text-white' : textColor
              } overflow-clip overflow-ellipsis w-full max-w-full`}
            >
              {log.message}
            </pre>
            <time
              dateTime={formatTimestamp(log.timeStamp)}
              className={`flex-none text-xs leading-5 ${
                selected ? 'text-gray-50' : 'text-gray-500'
              }`}
            >
              {formatTimestamp(log.timeStamp)}
            </time>
          </li>
        ))}
      </ul>
    </>
  )
}
