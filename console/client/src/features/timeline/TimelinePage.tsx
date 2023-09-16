import { ListBulletIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { PageHeader } from '../../components/PageHeader'
import { EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { Timeline } from './Timeline'
import { TimelineFilterPanel } from './filters/TimelineFilterPanel'
import { TimeSettings, TimelineTimeControls } from './filters/TimelineTimeControls'

export const TimelinePage = () => {
  const [timeSettings, setTimeSettings] = React.useState<TimeSettings>({ isTailing: true, isPaused: false })
  const [filters, setFilters] = React.useState<EventsQuery_Filter[]>([])

  const handleTimeSettingsChanged = (settings: TimeSettings) => {
    setTimeSettings(settings)
  }

  const handleFiltersChanged = (filters: EventsQuery_Filter[]) => {
    setFilters(filters)
  }

  return (
    <>
      <PageHeader icon={<ListBulletIcon />} title='Events'>
        <TimelineTimeControls onTimeSettingsChange={handleTimeSettingsChanged} />
      </PageHeader>
      <div className='flex h-full'>
        <TimelineFilterPanel onFiltersChanged={handleFiltersChanged} />
        <div className='flex-grow'>
          <Timeline timeSettings={timeSettings} filters={filters} />
        </div>
      </div>
    </>
  )
}
