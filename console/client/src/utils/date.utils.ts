import {Duration, Timestamp} from '@bufbuild/protobuf'

export function formatTimestamp(timestamp?: Timestamp): string {
  if (!timestamp) return '(no date)'
  return timestamp.toDate().toLocaleString()
}

export function formatDuration(duration?: Duration): string {
  if (!duration) return '(no duration)'
  return (duration.seconds * BigInt(1000) + BigInt(duration.nanos) / BigInt(1000000)) + 'ms'
}
