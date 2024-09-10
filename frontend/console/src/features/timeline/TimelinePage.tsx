import { ListViewIcon } from 'hugeicons-react'
import { useSearchParams } from 'react-router-dom'
import { Page } from '../../layout'
import type { EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { Timeline } from './Timeline'
import { TimelineFilterPanel } from './filters/TimelineFilterPanel'
import { TIME_RANGES, type TimeSettings, TimelineTimeControls } from './filters/TimelineTimeControls'
import { TimelineUrlState } from '../../api/timeline/timeline-url-state'
import { useEffect, useState } from 'react'

export const TimelinePage = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const [state] = useState(new TimelineUrlState(searchParams))

  console.warn('searchParams', searchParams.toString())
  console.warn('state', state.getSearchParams().toString())

  const timeSettings = state.time
  const filters = state.filters

  // const [timeSettings, setTimeSettings] = useState<TimeSettings>(params.time)
  // const [filters, setFilters] = useState<EventsQuery_Filter[]>([])
  // const [selectedTimeRange, setSelectedTimeRange] = useState(TIME_RANGES.tail)
  // const [isTimelinePaused, setIsTimelinePaused] = useState(false)
  // const initialEventId = searchParams.get('id')
  // useEffect(() => {
  //   if (initialEventId) {
  //     // if we're loading a specific event, we don't want to tail.
  //     setSelectedTimeRange(TIME_RANGES['5m'])
  //     setIsTimelinePaused(true)
  //   }
  // }, [])

  const handleTimeSettingsChanged = (settings: TimeSettings) => {
    // setTimeSettings(settings)
    state.time = settings
    setSearchParams(state.getSearchParams())
  }

  const handleFiltersChanged = (filters: EventsQuery_Filter[]) => {
    // setFilters(filters)
    state.filters = filters
    console.log('params.filters', JSON.stringify(state.filters))
    setSearchParams(state.getSearchParams())
  }

  return (
    <SidePanelProvider>
      <Page>
        <Page.Header icon={<ListViewIcon className='size-5' />} title='Events'>
          <TimelineTimeControls selectedTimeRange={state.timeRange} isTimelinePaused={state.time.isPaused} onTimeSettingsChange={handleTimeSettingsChanged} />
        </Page.Header>
        <Page.Body className='flex'>
          <div className='sticky top-0 flex-none overflow-y-auto'>
            <TimelineFilterPanel onFiltersChanged={handleFiltersChanged} />
          </div>
          <div className='flex-grow overflow-y-scroll'>
            <Timeline timeSettings={timeSettings} filters={filters} />
          </div>
        </Page.Body>
      </Page>
    </SidePanelProvider>
  )
}
