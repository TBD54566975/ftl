import type { AsyncExecuteEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { refString } from '../verbs/verb.utils'

export const TimelineAsyncExecute = ({ asyncExecute }: { asyncExecute: AsyncExecuteEvent }) => {
  const verbRef = (asyncExecute.verbRef?.module && refString(asyncExecute.verbRef)) || 'unknown'
  const title = `Async execution of verb ${verbRef}`
  return (
    <span title={title}>
      {'Async execution of verb '}
      <span className='text-indigo-500 dark:text-indigo-300'>{verbRef}</span>
    </span>
  )
}
