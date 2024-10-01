import type { CronScheduledEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatTimestampShort } from '../../utils/date.utils.ts'
import { verbRefString } from '../verbs/verb.utils'

export const TimelineCronScheduled = ({ cron }: { cron: CronScheduledEvent }) => {
  const title = `Cron (${cron.schedule}) scheduled at ${cron.scheduledAt}`
  return (
    <span title={title}>
      {'Cron '}
      <span className='text-indigo-500 dark:text-indigo-300'>{cron.schedule} </span>
      {'verb '}
      {cron.verbRef?.module && <span className='text-indigo-500 dark:text-indigo-300'>{verbRefString(cron.verbRef)} </span>}
      {'scheduled for '}
      <span className='text-indigo-500 dark:text-indigo-300'>{formatTimestampShort(cron.scheduledAt)}</span>
    </span>
  )
}
