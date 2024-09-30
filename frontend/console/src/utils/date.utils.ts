import type { Duration, Timestamp } from '@bufbuild/protobuf'

export const formatTimestamp = (timestamp?: Timestamp): string => {
  if (!timestamp) return '(no date)'
  return timestamp.toDate().toLocaleString()
}

export const formatDuration = (duration?: Duration): string => {
  if (!duration) return '(no duration)'
  return `${duration.seconds * BigInt(1000) + BigInt(duration.nanos) / BigInt(1000000)}ms`
}

export const formatTimestampShort = (timestamp?: Timestamp): string => {
  if (!timestamp) return '(no date)'
  const date = timestamp.toDate()
  const month = date.toLocaleString('default', { month: 'short' })

  const formattedDate =
    `${month} ${String(date.getDate()).padStart(2, '0')} ` +
    `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}:${String(date.getSeconds()).padStart(2, '0')}` +
    `.${String(date.getMilliseconds()).padStart(3, '0')}`

  return formattedDate
}

export const formatTimestampTime = (timestamp?: Timestamp): string => {
  if (!timestamp) return '(no date)'
  const date = timestamp.toDate()

  const formattedDate = `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(
    2,
    '0',
  )}:${String(date.getSeconds()).padStart(2, '0')}`

  return formattedDate
}

export const compareTimestamps = (a: Timestamp | undefined, b: Timestamp | undefined): number => {
  const timeASeconds = a?.seconds || 0n
  const timeBSeconds = b?.seconds || 0n

  // First, compare the seconds
  if (timeASeconds !== timeBSeconds) {
    return Number(timeASeconds - timeBSeconds)
  }

  // If seconds are identical, compare the nanoseconds
  const timeANanos = a?.nanos || 0
  const timeBNanos = b?.nanos || 0

  return timeANanos - timeBNanos
}
