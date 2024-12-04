import {
  Call02Icon,
  CallIncoming04Icon,
  CustomerServiceIcon,
  Menu01Icon,
  PackageReceiveIcon,
  Rocket01Icon,
  Satellite03Icon,
  SatelliteIcon,
  TimeQuarterPassIcon,
} from 'hugeicons-react'
import type { Event } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
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
        return <TimeQuarterPassIcon className={`${style} ${textColor}`} />
      case 'asyncExecute':
        return <CustomerServiceIcon className={`${style} ${textColor}`} />
      case 'pubsubPublish':
        return <SatelliteIcon className={`${style} ${textColor}`} />
      case 'pubsubConsume':
        return <Satellite03Icon className={`${style} ${textColor}`} />
      default:
        return <Menu01Icon className={`${style}`} />
    }
  }

  return <div>{icon(event)}</div>
}
