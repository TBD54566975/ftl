import { ListViewIcon } from 'hugeicons-react'
import { Page } from '../../layout'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { useTimelineState } from '../../api/timeline/use-timeline-state'
import { Timeline } from './Timeline'
import { TimelineFilterPanel } from './filters/TimelineFilterPanel'

export const TimelinePage = () => {
  // const [searchParams, setSearchParams] = useSearchParams()
  // const [state] = useState(new TimelineUrlState(searchParams))

  // console.warn('searchParams', searchParams.toString())
  // console.warn('state', state.getSearchParams().toString())

  // const timeSettings = state.time
  // const filters = state.filters

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

  // const handleTimeSettingsChanged = (settings: TimeSettings) => {
  //   // setTimeSettings(settings)
  //   state.time = settings
  //   setSearchParams(state.getSearchParams())
  // }

  const handleFiltersChanged = (filters: EventsQuery_Filter[]) => {
    //   // setFilters(filters)
    //   state.filters = filters
    //   console.log('params.filters', JSON.stringify(state.filters))
    //   setSearchParams(state.getSearchParams())
  }

  const [timelineState, _] = useTimelineState()

  return (
    <SidePanelProvider>
      <Page>
        {JSON.stringify(timelineState)}
        <Page.Header icon={<ListViewIcon className='size-5' />} title='Events'>
          {//<TimelineTimeControls selectedTimeRange={state.timeRange} isTimelinePaused={state.time.isPaused} onTimeSettingsChange={handleTimeSettingsChanged} />
          }
        </Page.Header>
        <Page.Body className='flex'>
          <div className='sticky top-0 flex-none overflow-y-auto'>
            <TimelineFilterPanel onFiltersChanged={handleFiltersChanged} />
          </div>
          <div className='flex-grow overflow-y-scroll'>
            <Timeline />
          </div>
        </Page.Body>
      </Page>
    </SidePanelProvider>
  )
}
