import { Timestamp } from '@bufbuild/protobuf'
import { useContext } from 'react'
import { AttributeBadge } from '../../../components/AttributeBadge'
import { CloseButton } from '../../../components/CloseButton'
import { DeploymentUpdatedEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../../providers/side-panel-provider'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { TimelineTimestamp } from './TimelineTimestamp'

export const TimelineDeploymentUpdatedDetails = ({
  event,
  deployment,
}: {
  event: Event
  deployment: DeploymentUpdatedEvent
}) => {
  const { closePanel } = useContext(SidePanelContext)

  return (
    <>
      <div className={`bg-blue-500 dark:bg-blue-300 h-2 w-full`}></div>
      <div className='p-4'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center space-x-2'>
            <span
              className={
                'text-blue-500 bg-blue-300/30 dark:text-blue-300 dark:bg-blue-700/30 inline-flex items-center rounded-md px-2 py-1 text-xs font-medium'
              }
            >
              Deployment Updated
            </span>
            <TimelineTimestamp timestamp={event.timeStamp ?? new Timestamp()} />
          </div>
          <CloseButton onClick={closePanel} />
        </div>

        <DeploymentCard className='mt-4' deploymentName={deployment.name} />

        <ul className='pt-4 space-y-2'>
          <li>
            <AttributeBadge name='Name' value={deployment.name} />
          </li>
          <li>
            <AttributeBadge name='MinReplicas' value={deployment.minReplicas.toString()} />
          </li>
          <li>
            <AttributeBadge name='PrevMinReplicas' value={deployment.prevMinReplicas.toString()} />
          </li>
        </ul>
      </div>
    </>
  )
}
