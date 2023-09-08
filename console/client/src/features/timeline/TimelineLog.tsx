import {LogEntry} from '../../protos/xyz/block/ftl/v1/console/console_pb'

type Props = {
  log: LogEntry
}

export const TimelineLog: React.FC<Props> = ({log}) => {
  return <span>{log.message}</span>
}
