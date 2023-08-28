import {Timestamp} from '@bufbuild/protobuf'
import React from 'react'
import {useClient} from '../../hooks/use-client.ts'
import {ConsoleService} from '../../protos/xyz/block/ftl/v1/console/console_connect.ts'
import {
  Call,
  Module,
  Verb,
} from '../../protos/xyz/block/ftl/v1/console/console_pb'
import {SidePanelContext} from '../../providers/side-panel-provider.tsx'
import {formatDuration, formatTimestamp} from '../../utils/date.utils.ts'
import {TimelineCallDetails} from '../timeline/details/TimelineCallDetails.tsx'

type Props = {
  module?: Module
  verb?: Verb
}

export const VerbCalls: React.FC<Props> = ({module, verb}) => {
  const client = useClient(ConsoleService)
  const [calls, setCalls] = React.useState<Call[]>([])
  const {openPanel} = React.useContext(SidePanelContext)

  React.useEffect(() => {
    const fetchCalls = async () => {
      const response = await client.getCalls({
        module: module?.name,
        verb: verb?.verb?.name,
      })
      setCalls(response.calls)
    }
    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    fetchCalls()
  }, [client, module, verb])

  const handleClick = (call: Call) => {
    openPanel(
      <TimelineCallDetails
        timestamp={call.timeStamp ?? new Timestamp()}
        call={call}
      />
    )
  }

  return (
    <>
      <div className='grid'>
        <table className='mt-6 w-full text-left'>
          <thead className='border-b border-white/10 text-sm leading-6 dark:text-white'>
            <tr>
              <th
                scope='col'
                className='hidden py-2 pl-0 pr-8 font-semibold sm:table-cell'
              >
                Request
              </th>
              <th
                scope='col'
                className='py-2 pl-4 pr-8 font-semibold sm:pl-6 lg:pl-8'
              >
                Source
              </th>
              <th
                scope='col'
                className='py-2 pl-0 pr-4 text-right font-semibold sm:pr-8 sm:text-left lg:pr-20'
              >
                Time
              </th>
              <th
                scope='col'
                className='hidden py-2 pr-0 text-right font-semibold md:table-cell'
              >
                Duration(ms)
              </th>
            </tr>
          </thead>
          <tbody className='divide-y divide-black/5 dark:divide-white/5'>
            {calls.map((call, index) => (
              <tr
                key={index}
                onClick={() => handleClick(call)}
              >
                <td className='hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8'>
                  <div className='flex gap-x-3'>
                    <div className='font-mono text-sm leading-6'>
                      {call.requestKey?.toString()}
                    </div>
                  </div>
                </td>
                <td className='py-4 pl-4 pr-8 sm:pl-6 lg:pl-8'>
                  <div className='flex items-center gap-x-4'>
                    <div className='truncate text-sm font-medium leading-6 text-gray-700 dark:text-white'>
                      {call.sourceVerbRef &&
                        [
                          call.sourceVerbRef.module,
                          call.sourceVerbRef.name,
                        ].join(':')}
                    </div>
                  </div>
                </td>
                <td className='hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8'>
                  <div className='flex gap-x-3'>
                    <div className='font-mono text-sm leading-6 text-gray-500 dark:text-gray-400'>
                      {formatTimestamp(call.timeStamp)}
                    </div>
                  </div>
                </td>
                <td
                  className={`hidden py-4 pr-0 text-right text-sm leading-6 text-gray-500 dark:text-gray-400 md:table-cell`}
                >
                  {formatDuration(call.duration)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  )
}
