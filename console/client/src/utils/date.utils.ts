export function dateFromTimestamp(timeStamp: bigint): string {
  return new Date(Number(timeStamp) * 1000).toLocaleString()
}

export function timeStampFromDate(date: Date): bigint {
  return BigInt(Math.floor(date.getTime() / 1000))
}
