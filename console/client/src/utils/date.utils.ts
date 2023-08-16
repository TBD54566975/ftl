import { Timestamp } from '@bufbuild/protobuf'

export function formatTimestamp(timestamp?: Timestamp): string {
  return timestamp?.toDate()?.toLocaleString() || '(no date)'
}
