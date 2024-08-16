import { callFilter } from './timeline-filters.ts'
import { useTimelineCalls } from './use-timeline-calls.ts'

export const useStreamVerbCalls = (moduleName?: string, verbName?: string, enabled = true) => {
  return useTimelineCalls(true, [callFilter(moduleName || '', verbName)], enabled)
}
