import { ListBulletIcon } from '@heroicons/react/24/outline'
import { PageHeader } from '../../components/PageHeader'
import { Timeline } from './Timeline'
import { TimelineFilterPanel } from './filters/TimelineFilterPanel'
import { TimelineTimeControls } from './filters/TimelineTimeControls'

export const TimelinePage = () => {
  return (
    <>
      <PageHeader icon={<ListBulletIcon />} title='Events'>
        <TimelineTimeControls />
      </PageHeader>
      <div className='flex h-full'>
        <TimelineFilterPanel />
        <div className='flex-grow'>
          <Timeline />
        </div>
      </div>
    </>
  )
}
