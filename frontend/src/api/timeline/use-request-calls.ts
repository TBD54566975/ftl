import { requestKeysFilter } from './timeline-filters'
import { useTimelineCalls } from './use-timeline-calls'

export const useRequestCalls = (requestKey?: string) => {
  return useTimelineCalls(true, [requestKeysFilter([requestKey || ''])], !!requestKey)
}
