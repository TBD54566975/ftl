import type { CronScheduledEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatTimestampShort } from '../../utils/date.utils.ts'
import { verbRefString } from '../verbs/verb.utils'

export const TimelineCronScheduled = ({ cron }: { cron: CronScheduledEvent }) => {
  const verbRef = cron.verbRef?.module && verbRefString(cron.verbRef) || 'unknown'
  const scheduledAt = formatTimestampShort(cron.scheduledAt)
  const title = `Cron ${cron.schedule} verb ${verbRef} scheduled for ${scheduledAt}`
  return (
    <span title={title}>
      {'Cron '}
      <span className='text-indigo-500 dark:text-indigo-300'>{cron.schedule}</span>
      {' verb '}
      <span className='text-indigo-500 dark:text-indigo-300'>{verbRef}</span>
      {' scheduled for '}
      <span className='text-indigo-500 dark:text-indigo-300'>{scheduledAt}</span>
    </span>
  )
}