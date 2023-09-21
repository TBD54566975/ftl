import { LogEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'

interface Props {
  log: LogEvent
}

export const TimelineLog = ({ log }: Props) => {
  return <span>{log.message}</span>
}
