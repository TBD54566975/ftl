export function dateFromTimestamp(timeStamp: bigint): string {
  return new Date(Number(timeStamp) * 1000).toLocaleString()
}
