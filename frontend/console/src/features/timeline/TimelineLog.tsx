import type { LogEvent } from '../../protos/xyz/block/ftl/v1/event_pb'

export const TimelineLog = ({ log }: { log: LogEvent }) => {
  return <span title={log.message}>{log.message}</span>
}
