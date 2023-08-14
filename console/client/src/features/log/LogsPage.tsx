import { useEffect, useState } from 'react'
import { useClient } from '../../hooks/use-client'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { LogEntry } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { dateFromTimestamp, timeStampFromDate } from '../../utils/date.utils'

export default function LogsPage() {
  const client = useClient(ConsoleService)
  const [ logs, setLogs ] = useState<LogEntry[]>([])

  useEffect(() => {
    const abortController = new AbortController()
    async function streamLogs() {
      const newLogs: LogEntry[] = []
      const afterTime = new Date()
      afterTime.setMinutes(afterTime.getMinutes() - 5)

      for await (const response of client.streamLogs(
        { afterTime: timeStampFromDate(afterTime) },
        { signal: abortController.signal })
      ) {
        if (response.log) {
          newLogs.push(response.log)
        }

        if (!response.more) {
          setLogs(newLogs)
        }
      }

    }
    streamLogs()
    return () => {
      abortController.abort()
    }
  }, [ client ])

  return (
    <>
      <h2 className='text-base font-semibold dark:text-white'>Logs</h2>
      <table className='mt-6 w-full whitespace-nowrap text-left'>
        <colgroup>
          <col className='w-full sm:w-4/12' />
          <col className='lg:w-4/12' />
          <col className='lg:w-2/12' />
          <col className='lg:w-1/12' />
          <col className='lg:w-1/12' />
        </colgroup>
        <thead className='border-b border-white/10 text-sm leading-6 dark:text-white'>
          <tr>
            <th
              scope='col'
              className='py-2 pl-0 pr-8 font-semibold'
            >
              Deployment
            </th>
            <th
              scope='col'
              className='hidden py-2 pl-0 pr-8 font-semibold sm:table-cell'
            >
              Runner
            </th>
            <th
              scope='col'
              className='py-2 pl-0 pr-4 text-right font-semibold sm:pr-8 sm:text-left lg:pr-20'
            >
              Level
            </th>
            <th
              scope='col'
              className='hidden py-2 pl-0 pr-8 font-semibold md:table-cell lg:pr-20'
            >
              Attributes
            </th>
            <th
              scope='col'
              className='hidden py-2 pl-0 pr-8 font-semibold md:table-cell lg:pr-20'
            >
              Message
            </th>
            <th
              scope='col'
              className='hidden py-2 pl-0 pr-8 font-semibold md:table-cell lg:pr-20'
            >
              Error
            </th>
            <th
              scope='col'
              className='hidden py-2 pl-0 pr-4 text-right font-semibold sm:table-cell sm:pr-6 lg:pr-8'
            >
              TimeStamp
            </th>
          </tr>
        </thead>
        <tbody className='divide-y divide-white/5'>
          {logs.map((log, index) => (
            <tr key={index}>
              <td className='hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8'>
                <div className='flex gap-x-3'>
                  <div className='font-mono text-sm leading-6 text-gray-400'>
                    {log.deploymentKey}
                  </div>
                </div>
              </td>
              <td className='hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8'>
                <div className='flex gap-x-3'>
                  <div className='font-mono text-sm leading-6 text-gray-400'>
                    {log.runnerKey}
                  </div>
                </div>
              </td>
              <td className='py-4 pl-0 pr-4 text-sm leading-6 sm:pr-8 lg:pr-20'>
                <div className='flex items-center justify-end gap-x-2 sm:justify-start'>
                  <div className={`rounded-md bg-gray-700/40 px-2 py-1 text-xs font-medium text-gray-400 ring-1 ring-inset ring-white/10`}>
                    {log.logLevel}
                  </div>
                </div>
              </td>
              <td className='hidden py-4 pl-0 pr-8 text-sm leading-6 text-gray-400 md:table-cell lg:pr-20'>
                <div className='flex gap-x-3'>
                  <div className='font-mono text-sm leading-6 text-gray-400'>
                    {JSON.stringify(log.attributes)}
                  </div>
                </div>
              </td>
              <td
                className={`hidden py-4 pl-0 pr-4 text-sm leading-6 text-gray-400 sm:table-cell sm:pr-6 lg:pr-8`}
              >
                <div className='truncate text-sm font-medium leading-6 dark:text-white'>{log.message}</div>
              </td>
              <td
                className={`hidden py-4 pl-0 pr-4 text-sm leading-6 text-gray-400 sm:table-cell sm:pr-6 lg:pr-8`}
              >
                <div className='truncate text-sm font-medium leading-6 dark:text-white'>{log.error}</div>
              </td>
              <td>
                <time
                  dateTime={dateFromTimestamp(log.timeStamp)}
                  className='flex-none py-0.5 text-xs leading-5 text-gray-500'
                >
                  {dateFromTimestamp(log.timeStamp)}
                </time>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  )
}
