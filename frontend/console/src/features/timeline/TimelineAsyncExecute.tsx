import type { AsyncExecuteEvent } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { refString } from '../modules/decls/verb/verb.utils'
import { asyncEventTypeString } from './timeline.utils'

export const TimelineAsyncExecute = ({ asyncExecute }: { asyncExecute: AsyncExecuteEvent }) => {
  const verbRef = (asyncExecute.verbRef?.module && refString(asyncExecute.verbRef)) || 'unknown'
  const title = `Async execution of verb ${verbRef}`
  return (
    <span title={title}>
      {'Async '}
      <span className='text-indigo-500 dark:text-indigo-300'>{asyncEventTypeString(asyncExecute.asyncEventType)}</span>
      {' execution of verb '}
      <span className='text-indigo-500 dark:text-indigo-300'>{verbRef}</span>
    </span>
  )
}
