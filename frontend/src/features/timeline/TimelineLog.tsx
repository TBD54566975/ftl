import type { LogEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'

export const TimelineLog = ({ log }: { log: LogEvent }) => {
  return <span>{log.message}</span>
}
