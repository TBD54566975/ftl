import { ListViewIcon } from 'hugeicons-react'
import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Page } from '../../layout'
import type { EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { Timeline } from './Timeline'
import { TimelineFilterPanel } from './filters/TimelineFilterPanel'
import { TIME_RANGES, type TimeSettings, TimelineTimeControls } from './filters/TimelineTimeControls'
import { newTimelineStateFromFilters, TimelineState } from '../../api/timeline/timeline-state'
import { useModules } from '../../api/modules/use-modules'

export const TimelinePage = () => {
  const modules = useModules()
  const [searchParams, setSearchParams] = useSearchParams()
  const timelineState = new TimelineState(searchParams, modules.data?.modules)
  console.log('--- timelinePage: timelineState', JSON.stringify(timelineState))
  const [timeSettings, setTimeSettings] = useState<TimeSettings>({ isTailing: timelineState.isTailing, isPaused: timelineState.isPaused })
  const [filters, setFilters] = useState<EventsQuery_Filter[]>(timelineState.getFilters())
  const [selectedTimeRange, setSelectedTimeRange] = useState(TIME_RANGES.tail) // TODO: timelineState.getTimeRange()
  const [isTimelinePaused, setIsTimelinePaused] = useState(timelineState.isPaused)

  useEffect(() => {
    console.warn('TimelinePage: searchParams changed', searchParams.toString())
  }, [searchParams])

  useEffect(() => {
    if (timelineState.eventId) {
      // if we're loading a specific event, we don't want to tail.
      setSelectedTimeRange(TIME_RANGES['5m'])
      setIsTimelinePaused(true)
    }
  }, [])

  useEffect(() => {
    console.log('yassssuuuuuuuuu TimelinePage: filters, timeSettings, isTimelinePaused changed', JSON.stringify(filters), JSON.stringify(timeSettings), isTimelinePaused)
    console.log('modules.data?.modules', modules.data?.modules)
    const timelineState = new TimelineState(searchParams, modules.data?.modules)
    timelineState.updateFromTimeSettings(timeSettings)
    timelineState.updateFromFilters(filters)
    timelineState.isPaused = isTimelinePaused
    setSearchParams(timelineState.getSearchParams())
  }, [filters, timeSettings, isTimelinePaused, modules.data?.modules])

  const handleTimeSettingsChanged = (settings: TimeSettings) => {
    setTimeSettings(settings)
  }

  const handleFiltersChanged = (filters: EventsQuery_Filter[]) => {
    console.log('got new filters', filters)
    setFilters(filters)
  }

  return (
    <SidePanelProvider>
      <Page>
        <Page.Header icon={<ListViewIcon className='size-5' />} title='Events'>
          <TimelineTimeControls selectedTimeRange={selectedTimeRange} isTimelinePaused={isTimelinePaused} onTimeSettingsChange={handleTimeSettingsChanged} />
        </Page.Header>
        <Page.Body className='flex'>
          <div className='sticky top-0 flex-none overflow-y-auto'>
            <TimelineFilterPanel onFiltersChanged={handleFiltersChanged} timelineState={timelineState} />
          </div>
          <div className='flex-grow overflow-y-scroll'>
            <Timeline timeSettings={timeSettings} filters={filters} />
          </div>
        </Page.Body>
      </Page>
    </SidePanelProvider>
  )
}
