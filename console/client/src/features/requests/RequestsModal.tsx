import {Dialog, Transition} from '@headlessui/react'
import {ChevronRightIcon} from '@heroicons/react/20/solid'
import React from 'react'
import {useLocation, useNavigate, useSearchParams} from 'react-router-dom'
import {useClient} from '../../hooks/use-client.ts'
import {ConsoleService} from '../../protos/xyz/block/ftl/v1/console/console_connect.ts'
import {Call} from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'
import {formatDuration, formatTimestamp} from '../../utils/date.utils.ts'

export function RequestModal() {
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams] = useSearchParams()
  const client = useClient(ConsoleService)
  const [calls, setCalls] = React.useState<Call[]>([])
  const key = searchParams.get('requests') ?? undefined
  const moduleName = searchParams.get('details')
  React.useEffect(() => {
    const fetchRequestCalls = async () => {
      if (key === undefined) {
        return
      }
      const response = await client.getRequestCalls({requestKey: key})
      setCalls(response.calls)
    }

    void fetchRequestCalls()
  }, [client, key])

  const isOpen = searchParams.has('requests')

  const handleClose = () => {
    searchParams.delete('requests')
    navigate({...location, search: searchParams.toString()})
  }

  return (
    <Transition
      appear
      show={isOpen}
      as={React.Fragment}
    >
      <Dialog
        onClose={handleClose}
        as='div'
        className='relative z-10'
      >
        <Transition.Child
          as={React.Fragment}
          enter='ease-out duration-300'
          enterFrom='opacity-0'
          enterTo='opacity-100'
          leave='ease-in duration-200'
          leaveFrom='opacity-100'
          leaveTo='opacity-0'
        >
          <div className='fixed inset-0 bg-black bg-opacity-25' />
        </Transition.Child>
        <div className='fixed inset-0 overflow-y-auto'>
          <div className='flex min-h-full items-center justify-center p-4 text-center'>
            <Transition.Child
              as={React.Fragment}
              enter='ease-out duration-300'
              enterFrom='opacity-0 scale-95'
              enterTo='opacity-100 scale-100'
              leave='ease-in duration-200'
              leaveFrom='opacity-100 scale-100'
              leaveTo='opacity-0 scale-95'
            >
              <Dialog.Panel
                className={`w-full max-w-7xl transform overflow-hidden rounded-2xl bg-white p-6 text-left align-middle shadow-xl transition-all`}
              >
                <Dialog.Title
                  as='h3'
                  className='text-lg font-medium leading-6 text-gray-900'
                >
                  <ol
                    role='list'
                    className='flex items-center space-x-4'
                  >
                    <li>
                      <div className='flex items-center'>
                        <button
                          className='focus:outline-none'
                          onClick={handleClose}
                        >
                          <span className='capitalize ml-4 text-sm font-medium text-gray-400 hover:text-gray-500'>
                            {moduleName} (module)
                          </span>
                        </button>
                      </div>
                    </li>
                    <li>
                      <div className='flex items-center'>
                        <ChevronRightIcon
                          className='h-5 w-5 flex-shrink-0 text-gray-400'
                          aria-hidden='true'
                        />
                        <span className='capitalize ml-4 text-sm font-medium text-gray-400 hover:text-gray-500'>
                          {key} (requests)
                        </span>
                      </div>
                    </li>
                  </ol>
                </Dialog.Title>
                <div className='min-w-0 flex-auto p-8'>
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
                          className='py-2 pl-4 pr-8 font-semibold sm:pl-6 lg:pl-8'
                        >
                          Destination
                        </th>
                        <th
                          scope='col'
                          className='py-2 pl-0 pr-4 text-right font-semibold sm:pr-8 sm:text-left lg:pr-20'
                        >
                          Time
                        </th>
                        <th
                          scope='col'
                          className='hidden py-2 pl-0 pr-8 font-semibold md:table-cell lg:pr-20'
                        >
                          Duration(ms)
                        </th>
                        <th
                          scope='col'
                          className='hidden py-2 pl-0 pr-4 text-right font-semibold sm:table-cell sm:pr-6 lg:pr-8'
                        >
                          Request
                        </th>
                        <th
                          scope='col'
                          className='hidden py-2 pl-0 pr-4 text-right font-semibold sm:table-cell sm:pr-6 lg:pr-8'
                        >
                          Response
                        </th>
                        <th
                          scope='col'
                          className='hidden py-2 pl-0 pr-4 text-right font-semibold sm:table-cell sm:pr-6 lg:pr-8'
                        >
                          Error
                        </th>
                      </tr>
                    </thead>
                    <tbody className='divide-y divide-black/5 dark:divide-white/5'>
                      {calls.map(call => (
                        <tr key={call.destinationVerbRef?.module}>
                          <td className='hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8'>
                            <div className='flex gap-x-3'>
                              <div className='font-mono text-sm leading-6 text-gray-500 dark:text-gray-400'>
                                {call.requestKey?.toString()}
                              </div>
                            </div>
                          </td>
                          <td className='py-4 pl-4 pr-8 sm:pl-6 lg:pl-8'>
                            <div className='flex items-center gap-x-4'>
                              <div className='truncate text-sm font-medium leading-6 text-gray-700 dark:text-white'>
                                {call.sourceVerbRef?.module &&
                                  [
                                    call.sourceVerbRef.module,
                                    call.sourceVerbRef.name,
                                  ].join(':')}
                              </div>
                            </div>
                          </td>
                          <td className='py-4 pl-4 pr-8 sm:pl-6 lg:pl-8'>
                            <div className='flex items-center gap-x-4'>
                              <div className='truncate text-sm font-medium leading-6 text-gray-700 dark:text-white'>
                                {call.destinationVerbRef?.module &&
                                  [
                                    call.destinationVerbRef.module,
                                    call.destinationVerbRef.name,
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
                            className={`hidden py-4 pl-0 pr-8 text-right text-sm leading-6 text-gray-500 dark:text-gray-400 md:table-cell lg:pr-20`}
                          >
                            {formatDuration(call.duration)}
                          </td>
                          <td
                            className={`hidden py-4 pl-0 pr-4 text-right text-sm leading-6 text-gray-400 sm:table-cell sm:pr-6 lg:pr-8`}
                          >
                            <code>{call.request}</code>
                          </td>
                          <td
                            className={`hidden py-4 pl-0 pr-4 text-right text-sm leading-6 text-gray-400 sm:table-cell sm:pr-6 lg:pr-8`}
                          >
                            <code>{call.response}</code>
                          </td>
                          <td
                            className={`hidden py-4 pl-0 pr-4 text-right text-sm leading-6 text-gray-400 sm:table-cell sm:pr-6 lg:pr-8`}
                          >
                            {call.error}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition>
  )
}
