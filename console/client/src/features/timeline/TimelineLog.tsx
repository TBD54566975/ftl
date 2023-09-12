import { LogEntry } from '../../protos/xyz/block/ftl/v1/console/console_pb'

interface Props {
  log: LogEntry
}

export const TimelineLog = ({ log }: Props) => {
  return <span>{log.message}</span>
}
