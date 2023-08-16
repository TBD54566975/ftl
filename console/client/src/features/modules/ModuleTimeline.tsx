import { Module, TimelineEntry } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import React, { useEffect, useState } from 'react'
import { useClient } from '../../hooks/use-client.ts'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect.ts'
import { RocketLaunchIcon } from '@heroicons/react/24/solid'
import { classNames } from '../../utils/react.utils.ts'
import { Link } from 'react-router-dom'
import { formatDuration, formatTimestamp } from '../../utils/date.utils.ts'

type Props = {
  module?: Module
}

export const ModuleTimeline: React.FC<Props> = ({ module }) => {
  const client = useClient(ConsoleService)
  const [ entries, setEntries ] = useState<TimelineEntry[]>([])

  useEffect(() => {
    const fetchTimeline = async () => {
      const response = await client.getTimeline({ module: module?.name })
      setEntries(response.entries)
    }
    fetchTimeline()
  }, [ client, module ])

  return (
    <>
      <ul role='list'
        className='space-y-6'
      >
        {entries.map((entry, activityItemIdx) => (
          <li key={entry.toJsonString()}
            className='relative flex gap-x-4'
          >
            <div
              className={classNames(
                activityItemIdx === entries.length - 1 ? 'h-6' : '-bottom-6',
                'absolute left-0 top-0 flex w-6 justify-center'
              )}
            >
              <div className='w-px bg-gray-200 dark:bg-gray-600' />
            </div>
            {entry.entry.case === 'call' ? (
              <>
                <div className='relative flex h-6 w-6 flex-none items-center justify-center bg-white dark:bg-slate-800'>
                  <div className='h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300' />
                </div>
                <p className='flex-auto py-0.5 text-xs leading-5 text-gray-500 dark:text-gray-400'>
                  <span className='text-indigo-700 dark:text-indigo-400 '>
                    <Link to={`/requests/${entry.entry.value?.requestKey}`}
                      className='focus:outline-none'
                    >
                      Called{' '}
                    </Link>
                  </span>
                  <span className='font-medium text-gray-900 dark:text-white'>
                    {entry.entry.value?.destModule}:{entry.entry.value?.destVerb}
                  </span>
                  {entry.entry.value.sourceVerb && (
                    <>
                      {' '}
                      from{' '}
                      <span className='font-medium text-gray-900 dark:text-white'>
                        {entry.entry.value?.sourceModule}:{entry.entry.value?.sourceVerb}
                      </span>
                    </>
                  )}{' '}
                  ({formatDuration(entry.entry.value.duration)}).
                </p>
                <time
                  dateTime={formatTimestamp(entry.timeStamp)}
                  className='flex-none py-0.5 text-xs leading-5 text-gray-500'
                >
                  {formatTimestamp(entry.timeStamp)}
                </time>
              </>
            ) : (
              <>
                <div className='relative flex h-6 w-6 flex-none items-center justify-center bg-white dark:bg-slate-800'>
                  <RocketLaunchIcon className='h-6 w-6 text-indigo-500'
                    aria-hidden='true'
                  />
                </div>
                <p className='flex-auto py-0.5 text-xs leading-5 text-gray-500 dark:text-gray-400'>
                  Deployed <span className={`font-medium text-gray-900 dark:text-white`}>{entry.entry.value?.name}</span>{' '}
                  for language{' '}
                  <span className='font-medium text-gray-900 dark:text-white'>{entry.entry.value?.language}</span>.
                </p>
                <time
                  dateTime={formatTimestamp(entry.timeStamp)}
                  className='flex-none py-0.5 text-xs leading-5 text-gray-500'
                >
                  {formatTimestamp(entry.timeStamp)}
                </time>
              </>
            )}
          </li>
        ))}
      </ul>
    </>
  )
}
