import { ArrowRightOnRectangleIcon } from '@heroicons/react/24/outline'
import { Link } from 'react-router-dom'
import { Call } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatDuration, formatTimestamp } from '../../utils/date.utils'
import { panelColor } from '../../utils/style.utils'

type Props = {
  call: Call
}

export const TimelineCall: React.FC<Props> = ({ call }) => {
  return (
    <>
      <div className={`relative flex h-6 w-6 flex-none items-center justify-center ${panelColor}`}>
        <ArrowRightOnRectangleIcon className='h-6 w-6 text-indigo-500'
          aria-hidden='true'
        />
      </div>
      <p className='flex-auto py-0.5 text-xs leading-5 text-gray-500 dark:text-gray-400'>
        {call.sourceVerbRef && (
          <>
            <div className={`inline-block rounded-md dark:bg-gray-700/40 px-2 py-1 mr-2 text-xs font-medium text-gray-500 dark:text-gray-400 ring-1 ring-inset ring-black/10 dark:ring-white/10`}>
              {call.destinationVerbRef?.module}
            </div>
            Verb <span className='font-medium text-gray-900 dark:text-white mr-1'>{call.destinationVerbRef?.name}</span>
          </>
        )}

        <span className='text-indigo-700 dark:text-indigo-400 mr-1'>
          <Link to={`/requests/${call.requestKey}`}
            className='focus:outline-none'
          >
            Called
          </Link>
        </span>


        {call.sourceVerbRef?.module && (
          <>
            from
            <span className='font-medium text-gray-900 dark:text-white mx-1'>
              {call.sourceVerbRef.module}:{call.sourceVerbRef.name}
            </span>
          </>
        )}
        Duration ({formatDuration(call.duration)}).
      </p>
      <time
        dateTime={formatTimestamp(call.timeStamp)}
        className='flex-none py-0.5 text-xs leading-5 text-gray-500'
      >
        {formatTimestamp(call.timeStamp)}
      </time>
    </>
  )
}
