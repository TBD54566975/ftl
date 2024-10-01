import { AlarmClockIcon, Call02Icon, CallIncoming04Icon, Menu01Icon, PackageReceiveIcon, Rocket01Icon } from 'hugeicons-react'
import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { LogLevelBadgeSmall } from '../logs/LogLevelBadgeSmall'
import { eventTextColor } from './timeline.utils'

export const TimelineIcon = ({ event }: { event: Event }) => {
  const icon = (event: Event) => {
    const style = 'h4 w-4'
    const textColor = eventTextColor(event)

    switch (event.entry.case) {
      case 'call': {
        return event.entry.value.sourceVerbRef ? <Call02Icon className={`${style} ${textColor}`} /> : <CallIncoming04Icon className={`${style} ${textColor}`} />
      }
      case 'deploymentCreated':
        return <Rocket01Icon className={`${style} ${textColor}`} />
      case 'deploymentUpdated':
        return <Rocket01Icon className={`${style} ${textColor}`} />
      case 'log':
        return <LogLevelBadgeSmall logLevel={event.entry.value.logLevel} />
      case 'ingress':
        return <PackageReceiveIcon className={`${style} ${textColor}`} />
      case 'cronScheduled':
        return <AlarmClockIcon className={`${style} ${textColor}`} />
      default:
        return <Menu01Icon className={`${style}`} />
    }
  }

  return <div>{icon(event)}</div>
}
