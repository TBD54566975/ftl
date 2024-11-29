import type { Event } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { logLevelBgColor } from '../../logs/log.utils'
import { eventBackgroundColor } from '../timeline.utils'

export const TimelineDetailsColorBar = ({ event }: { event: Event }) => {
  let barColor = eventBackgroundColor(event)
  if (event.entry.case === 'log') {
    barColor = logLevelBgColor[event.entry.value.logLevel]
  }
  return <div className={`${barColor} h-2 w-full`} />
}
