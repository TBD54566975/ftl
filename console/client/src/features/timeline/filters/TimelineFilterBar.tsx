import { panelColor } from '../../../utils/style.utils'
import { LogLevelsFilter } from './LogLevelsFilter'
import { TimeFilter } from './TimeFilter'

interface Props {
  selectedEventTypes: string[]
  onEventTypesChanged: (eventType: string, checked: boolean) => void
  selectedLogLevels: number[]
  onLogLevelsChanged: (logLevel: number, checked: boolean) => void
  selectedTimeRange: string
  onSelectedTimeRangeChanged: (range: string) => void
}

export const TimelineFilterBar = ({
  selectedEventTypes,
  onEventTypesChanged,
  selectedLogLevels,
  onLogLevelsChanged,
  selectedTimeRange,
  onSelectedTimeRangeChanged,
}: Props) => {
  const eventButtonStyles = `relative inline-flex items-center px-2 py-1 text-xs font-semibold ring-1 ring-inset focus:z-10`

  const eventSelectedStyles = (eventType: string) => {
    return selectedEventTypes.includes(eventType)
      ? 'bg-indigo-600 text-white ring-indigo-200 dark:ring-indigo-400'
      : 'bg-white text-gray-700 ring-indigo-200 dark:ring-indigo-400'
  }

  const toggleEventType = (eventType: string) => {
    onEventTypesChanged(eventType, !selectedEventTypes.includes(eventType))
  }

  return (
    <>
      <div className={`sticky top-0 z-10 ${panelColor} shadow`}>
        <div className='flex items-center justify-between p-4'>
          <span className='isolate inline-flex rounded-md shadow-sm'>
            <button
              type='button'
              className={`${eventButtonStyles} ${eventSelectedStyles('log')} rounded-l-md`}
              onClick={() => toggleEventType('log')}
            >
              Logs
            </button>
            <button
              type='button'
              className={`${eventButtonStyles} ${eventSelectedStyles('call')} -ml-px`}
              onClick={() => toggleEventType('call')}
            >
              Calls
            </button>
            <button
              type='button'
              className={`${eventButtonStyles} ${eventSelectedStyles('deployment')} -ml-px rounded-r-md`}
              onClick={() => toggleEventType('deployment')}
            >
              Deployments
            </button>
          </span>
          <div className='flex items-center space-x-4'>
            <TimeFilter selectedRange={selectedTimeRange} onSelectedRangeChanged={onSelectedTimeRangeChanged} />
            <LogLevelsFilter selectedLogLevels={selectedLogLevels} onLogLevelsChanged={onLogLevelsChanged} />
          </div>
        </div>
      </div>
    </>
  )
}
