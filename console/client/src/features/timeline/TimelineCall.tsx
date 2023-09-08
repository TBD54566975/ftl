import {Call} from '../../protos/xyz/block/ftl/v1/console/console_pb'
import {verbRefString} from '../verbs/verb.utils'

type Props = {
  call: Call
}

export const TimelineCall: React.FC<Props> = ({call}) => {
  return (
    <span>
      {call.sourceVerbRef?.module && (
        <>
          <span className='text-indigo-500 dark:text-indigo-300'>
            {verbRefString(call.sourceVerbRef)}
          </span>
          {' -> '}
        </>
      )}
      {call.destinationVerbRef?.module && (
        <span className='text-indigo-500 dark:text-indigo-300'>
          {verbRefString(call.destinationVerbRef)}
        </span>
      )}
    </span>
  )
}
