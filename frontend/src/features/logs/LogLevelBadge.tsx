import { logLevelBadge, logLevelText } from './log.utils'

export const LogLevelBadge = ({ logLevel }: { logLevel: number }) => {
  return (
    <span
      className={`${logLevelBadge[logLevel]} inline-flex items-center rounded-md px-2 py-1 text-xs font-medium font-roboto-mono`}
    >
      {logLevelText[logLevel]}
    </span>
  )
}
