import { useContext, useState } from 'react'
import type { CallEvent, Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../providers/side-panel-provider'
import { formatDuration, formatTimestampShort } from '../../utils'
import { TimelineCallDetails } from '../timeline/details/TimelineCallDetails'
import { verbRefString } from '../verbs/verb.utils'

export const CallList = ({ calls }: { calls: Event[] | undefined }) => {
  const { openPanel, closePanel } = useContext(SidePanelContext)
  const [selectedCallId, setSelectedCallId] = useState<bigint | undefined>()

  const handleCallClicked = (event: Event) => {
    if (selectedCallId === event.id) {
      setSelectedCallId(undefined)
      closePanel()
      return
    }
    setSelectedCallId(event.id)
    const call = event.entry.value as CallEvent
    openPanel(<TimelineCallDetails timestamp={call.timeStamp} event={event} />)
  }

  return (
    <div className='flex flex-col h-full relative min-w-[710px]'>
      <div className='border border-gray-100 dark:border-slate-700 rounded h-full absolute inset-0'>
        <table className={'w-full table-fixed text-gray-600 dark:text-gray-300'}>
          <thead>
            <tr className='flex text-xs'>
              <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none sticky top-0 z-10'>Date</th>
              <th className='p-1 text-left border-b w-14 border-gray-100 dark:border-slate-700 flex-none sticky top-0 z-10'>Dur.</th>
              <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none sticky top-0 z-10'>Source</th>
              <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none sticky top-0 z-10'>Destination</th>
              <th className='p-1 text-left border-b border-gray-100 dark:border-slate-700 min-w-[60px] flex-1 flex-grow sticky top-0 z-10'>Request</th>
              <th className='p-1 text-left border-b border-gray-100 dark:border-slate-700 min-w-[70px] flex-1 flex-grow sticky top-0 z-10'>Response</th>
              <th className='p-1 text-left border-b border-gray-100 dark:border-slate-700 flex-1 flex-grow sticky top-0 z-10'>Error</th>
            </tr>
          </thead>
        </table>
        <div className='overflow-y-auto h-[calc(100%-1.5rem)]'>
          <table className={'w-full table-fixed text-gray-600 dark:text-gray-300'}>
            <tbody className='text-xs'>
              {calls?.map((callEvent, index) => {
                const call = callEvent.entry.value as CallEvent
                return (
                  <tr
                    key={`${index}-${callEvent.timeStamp?.toDate().toUTCString()}`}
                    className={`border-b border-gray-100 dark:border-slate-700 font-roboto-mono
                   ${selectedCallId === callEvent.id ? 'bg-indigo-50 dark:bg-slate-700' : ''} relative flex cursor-pointer hover:bg-indigo-50 dark:hover:bg-slate-700`}
                    onClick={() => handleCallClicked(callEvent)}
                  >
                    <td className='p-1 w-40 items-center flex-none text-gray-400 dark:text-gray-400'>{formatTimestampShort(callEvent.timeStamp)}</td>
                    <td className='p-1 w-14 items-center flex-none text-gray-400 dark:text-gray-400 truncate'>{formatDuration(call.duration)}</td>
                    <td
                      className='p-1 w-40 flex-none text-indigo-500 dark:text-indigo-300 truncate'
                      title={call.sourceVerbRef && verbRefString(call.sourceVerbRef)}
                    >
                      {call.sourceVerbRef && verbRefString(call.sourceVerbRef)}
                    </td>
                    <td
                      className='p-1 w-40 flex-none text-indigo-500 dark:text-indigo-300 truncate'
                      title={call.destinationVerbRef && verbRefString(call.destinationVerbRef)}
                    >
                      {call.destinationVerbRef && verbRefString(call.destinationVerbRef)}
                    </td>
                    <td className='p-1 min-w-[60px] flex-1 flex-grow truncate' title={call.request}>
                      {call.request}
                    </td>
                    <td className='p-1 min-w-[70px] flex-1 flex-grow truncate' title={call.response}>
                      {call.response}
                    </td>
                    <td className='p-1 flex-1 flex-grow truncate text-red-500' title={call.error}>
                      {call.error}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
