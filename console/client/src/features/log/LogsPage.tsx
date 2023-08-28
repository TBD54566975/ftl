import {Timestamp} from '@bufbuild/protobuf'
import React, {useEffect, useState} from 'react'
import {useClient} from '../../hooks/use-client'
import {ConsoleService} from '../../protos/xyz/block/ftl/v1/console/console_connect'
import {LogEntry} from '../../protos/xyz/block/ftl/v1/console/console_pb'
import {formatTimestampShort} from '../../utils/date.utils'
import {classNames} from '../../utils/react.utils'
import {logLevelBadge, logLevelText} from '../../utils/style.utils'

export default function LogsPage() {
  const client = useClient(ConsoleService)
  const [expandedLog, setExpandedLog] = useState<number | null>(null)
  const [logs, setLogs] = useState<LogEntry[]>([])

  useEffect(() => {
    const abortController = new AbortController()

    async function streamLogs() {
      const afterTime = new Date()
      afterTime.setHours(afterTime.getHours() - 1)

      for await (const response of client.streamLogs(
        {afterTime: Timestamp.fromDate(afterTime)},
        {signal: abortController.signal}
      )) {
        if (response.log) {
          setLogs(prevLogs => [response.log!, ...prevLogs])
        }
      }
    }

    void streamLogs()
    return () => {
      abortController.abort()
    }
  }, [client])

  return (
    <>
      <div className='flow-root'>
        <div className='-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8'>
          <div className='inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8'>
            <table className='min-w-full divide-y divide-gray-300 dark:divide-gray-400'>
              <thead>
                <tr>
                  <th
                    scope='col'
                    className='whitespace-nowrap px-2 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-100'
                  >
                    Level
                  </th>
                  <th
                    scope='col'
                    className='whitespace-nowrap px-2 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-100'
                  >
                    Date
                  </th>
                  <th
                    scope='col'
                    className='whitespace-nowrap px-2 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-100'
                  >
                    Message
                  </th>
                </tr>
              </thead>
              <tbody className='divide-y divide-gray-200 dark:divide-gray-600 bg-white dark:bg-slate-800'>
                {logs.map((log, index) => (
                  <React.Fragment key={index}>
                    <tr
                      onClick={() =>
                        setExpandedLog(expandedLog !== index ? index : null)
                      }
                    >
                      <td className='whitespace-nowrap px-2 py-2'>
                        <span
                          className={classNames(
                            `${logLevelBadge[log.logLevel]}`,
                            'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium text-gray-600'
                          )}
                        >
                          {logLevelText[log.logLevel]}
                        </span>
                      </td>
                      <td className='whitespace-nowrap px-2 py-2 text-sm '>
                        <time
                          dateTime={formatTimestampShort(log.timeStamp)}
                          className='flex-none py-0.5 text-xs leading-5 text-gray-500 dark:text-gray-400'
                        >
                          {formatTimestampShort(log.timeStamp)}
                        </time>
                      </td>
                      <td className='whitespace-nowrap px-2 py-2 text-sm text-gray-500 dark:text-gray-300'>
                        {log.message}
                      </td>
                    </tr>
                    {expandedLog === index && (
                      <tr>
                        <td colSpan={4}>
                          <div className='p-4 text-sm bg-white text-gray-600 dark:bg-slate-800 dark:text-gray-400'>
                            <div className='mb-2'>
                              <strong>Deployment Key:</strong>{' '}
                              {log.deploymentName}
                            </div>

                            {log.requestKey && (
                              <div className='mb-2'>
                                <strong>Request Key:</strong> {log.requestKey}
                              </div>
                            )}

                            <div className='mb-2'>
                              <strong>Attributes:</strong>
                              <div className='ml-4'>
                                {Object.entries(log.attributes).map(
                                  ([key, value]) => (
                                    <div key={key}>
                                      {key}: {value}
                                    </div>
                                  )
                                )}
                              </div>
                            </div>

                            {log.error && (
                              <div className='mt-2 text-red-500'>
                                <strong>Error:</strong> {log.error}
                              </div>
                            )}
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </>
  )
}
