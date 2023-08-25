import {ArrowRightOnRectangleIcon} from '@heroicons/react/24/outline'
import {Call} from '../../protos/xyz/block/ftl/v1/console/console_pb'
import {formatDuration, formatTimestamp} from '../../utils/date.utils'
import {panelColor, textColor} from '../../utils/style.utils'
import {classNames} from '../../utils/react.utils'
type Props = {
  call: Call
  selected?: boolean
}

export const TimelineCall: React.FC<Props> = ({call, selected}) => {
  return (
    <>
      <div
        className={`relative flex h-6 w-6 flex-none items-center justify-center ${panelColor}`}>
        <ArrowRightOnRectangleIcon
          className='h-6 w-6 text-indigo-500'
          aria-hidden='true'
        />
      </div>
      <div
        className={classNames(
          `relative flex gap-x-4 flex-auto w-full max-w-full p-1.5 `,
          selected && 'bg-indigo-600 rounded-md'
        )}>
        <div
          className={`flex-auto text-xs leading-5 ${
            selected ? 'text-gray-50' : textColor
          }`}>
          {call.destinationVerbRef && (
            <div
              className={`inline-block rounded-md dark:bg-gray-700/40 px-2 py-1 mr-1 text-xs font-medium ${
                selected ? 'text-white' : 'text-gray-500 dark:text-gray-400'
              } ring-1 ring-inset ring-black/10 dark:ring-white/10`}>
              {call.destinationVerbRef?.module}:{call.destinationVerbRef?.name}
            </div>
          )}
          <span className=' mr-1'>called</span>
          {call.sourceVerbRef?.module && (
            <>
              from
              <div
                className={`inline-block rounded-md dark:bg-gray-700/40 px-2 py-1 ml-1 mr-1 text-xs font-medium ${
                  selected ? 'text-white' : 'text-gray-500 dark:text-gray-400'
                } ring-1 ring-inset ring-black/10 dark:ring-white/10`}>
                {call.sourceVerbRef?.module}:{call.sourceVerbRef?.name}
              </div>
            </>
          )}
          ({formatDuration(call.duration)}).
        </div>
        <time
          dateTime={formatTimestamp(call.timeStamp)}
          className={`flex-none text-xs leading-5 ${
            selected ? 'text-gray-50' : 'text-gray-500'
          }`}>
          {formatTimestamp(call.timeStamp)}
        </time>
      </div>
    </>
  )
}
