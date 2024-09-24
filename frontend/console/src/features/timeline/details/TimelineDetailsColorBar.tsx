import type { Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { eventBackgroundColor } from '../timeline.utils'

export const TimelineDetailsColorBar = ({ event }: { event: Event }) => {
  const barColor = eventBackgroundColor(event)
  return <div className={`${barColor} h-2 w-full`} />
}
