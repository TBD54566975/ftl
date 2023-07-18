import { Call, Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import React, { useEffect, useState } from 'react'
import { useClient } from '../../hooks/use-client.ts'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect.ts'

type Props = {
  module?: Module
  verb?: Verb
}

export const VerbCalls: React.FC<Props> = ({ module, verb }) => {
  const client = useClient(ConsoleService)
  const [calls, setCalls] = useState<Call[]>([])

  useEffect(() => {
    const fetchCalls = async () => {
      const response = await client.getCalls({ module: module?.name, verb: verb?.verb?.name })
      setCalls(response.calls)
    }
    fetchCalls()
  }, [client, module, verb])

  return (
    <>
      <div className="grid">
        <table className="mt-6 w-full text-left">
          <thead className="border-b border-white/10 text-sm leading-6 dark:text-white">
            <tr>
              <th scope="col" className="py-2 pl-4 pr-8 font-semibold sm:pl-6 lg:pl-8">
                Source
              </th>
              <th scope="col" className="hidden py-2 pl-0 pr-8 font-semibold sm:table-cell">
                Request
              </th>
              <th scope="col" className="py-2 pl-0 pr-4 text-right font-semibold sm:pr-8 sm:text-left lg:pr-20">
                Time
              </th>
              <th scope="col" className="hidden py-2 pl-0 pr-8 font-semibold md:table-cell lg:pr-20">
                Duration(ms)
              </th>
              <th scope="col" className="hidden py-2 pl-0 pr-4 text-right font-semibold sm:table-cell sm:pr-6 lg:pr-8">
                Request
              </th>
              <th scope="col" className="hidden py-2 pl-0 pr-4 text-right font-semibold sm:table-cell sm:pr-6 lg:pr-8">
                Response
              </th>
              <th scope="col" className="hidden py-2 pl-0 pr-4 text-right font-semibold sm:table-cell sm:pr-6 lg:pr-8">
                Error
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-black/5 dark:divide-white/5">
            {calls.map(call => (
              <tr key={call.destModule}>
                <td className="py-4 pl-4 pr-8 sm:pl-6 lg:pl-8">
                  <div className="flex items-center gap-x-4">
                    <div className="truncate text-sm font-medium leading-6 text-gray-700 dark:text-white">
                      {call.sourceModule.length > 0 && [call.sourceModule, call.sourceVerb].join(':')}
                    </div>
                  </div>
                </td>
                <td className="hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8">
                  <div className="flex gap-x-3">
                    <div className="font-mono text-sm leading-6 text-gray-500 dark:text-gray-400">
                      {call.requestId.toString()}
                    </div>
                  </div>
                </td>
                <td className="hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8">
                  <div className="flex gap-x-3">
                    <div className="font-mono text-sm leading-6 text-gray-500 dark:text-gray-400">
                      {new Date(Number(call.timeStamp)).toLocaleString()}
                    </div>
                  </div>
                </td>
                <td className="hidden py-4 pl-0 pr-8 text-right text-sm leading-6 text-gray-500 dark:text-gray-400 md:table-cell lg:pr-20">
                  {call.durationMs.toString()}
                </td>
                <td className="hidden py-4 pl-0 pr-4 text-right text-sm leading-6 text-gray-400 sm:table-cell sm:pr-6 lg:pr-8">
                  <code>{call.request}</code>
                </td>
                <td className="hidden py-4 pl-0 pr-4 text-right text-sm leading-6 text-gray-400 sm:table-cell sm:pr-6 lg:pr-8">
                  <code>{call.response}</code>
                </td>
                <td className="hidden py-4 pl-0 pr-4 text-right text-sm leading-6 text-gray-400 sm:table-cell sm:pr-6 lg:pr-8">
                  {call.error}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  )
}
