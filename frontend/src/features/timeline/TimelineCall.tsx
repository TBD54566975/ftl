import type { CallEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { verbRefString } from '../verbs/verb.utils'

export const TimelineCall = ({ call }: { call: CallEvent }) => {
  const title = `${call.sourceVerbRef?.module ? `${verbRefString(call.sourceVerbRef)} -> ` : ''}${verbRefString(call.destinationVerbRef)}`;
  return (
    <span title={title}>
      {call.sourceVerbRef?.module && (
        <>
          <span className='text-indigo-500 dark:text-indigo-300'>{verbRefString(call.sourceVerbRef)}</span>
          {' -> '}
        </>
      )}
      {call.destinationVerbRef?.module && <span className='text-indigo-500 dark:text-indigo-300'>{verbRefString(call.destinationVerbRef)}</span>}
    </span>
  )
}
