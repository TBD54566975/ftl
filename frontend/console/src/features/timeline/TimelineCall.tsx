import type { CallEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { refString } from '../modules/decls/verb/verb.utils'

export const TimelineCall = ({ call }: { call: CallEvent }) => {
  const title = `${call.sourceVerbRef?.module ? `${refString(call.sourceVerbRef)} -> ` : ''}${call.destinationVerbRef ? refString(call.destinationVerbRef) : ''}`
  return (
    <span title={title}>
      {call.sourceVerbRef?.module && (
        <>
          <span className='text-indigo-500 dark:text-indigo-300'>{refString(call.sourceVerbRef)}</span>
          {' -> '}
        </>
      )}
      {call.destinationVerbRef?.module && <span className='text-indigo-500 dark:text-indigo-300'>{refString(call.destinationVerbRef)}</span>}
    </span>
  )
}
