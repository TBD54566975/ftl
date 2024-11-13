import type { LogEvent } from '../../protos/xyz/block/ftl/console/v1/console_pb'

export const TimelineLog = ({ log }: { log: LogEvent }) => {
  return <span title={log.message}>{log.message}</span>
}
