import { Duration, Timestamp } from '@bufbuild/protobuf'

export function formatTimestamp(timestamp?: Timestamp): string {
  if (!timestamp) return '(no date)'
  return timestamp.toDate().toLocaleString()
}

export function formatDuration(duration?: Duration): string {
  if (!duration) return '(no duration)'
  return duration.seconds * BigInt(1000) + BigInt(duration.nanos) / BigInt(1000000) + 'ms'
}

export function formatTimestampShort(timestamp?: Timestamp): string {
  if (!timestamp) return '(no date)'
  const date = timestamp.toDate()
  const month = date.toLocaleString('default', { month: 'short' })

  const formattedDate =
    `${month} ${String(date.getDate()).padStart(2, '0')}, ` +
    `${date.getHours() % 12 || 12}:${String(date.getMinutes()).padStart(2, '0')}:${String(date.getSeconds()).padStart(
      2,
      '0'
    )}` +
    `.${String(date.getMilliseconds()).padStart(3, '0')} ` +
    `${date.getHours() < 12 ? 'AM' : 'PM'}`

  return formattedDate
}
