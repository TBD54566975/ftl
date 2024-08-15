import { logLevelBadge, logLevelCharacter } from './log.utils'

export const LogLevelBadgeSmall = ({ logLevel }: { logLevel: number }) => {
  return (
    <span className={`flex rounded justify-center items-center pb-0.5 h-4 w-4 text-xs font-roboto-mono ${logLevelBadge[logLevel]}`}>
      {`${logLevelCharacter[logLevel]}`}
    </span>
  )
}
