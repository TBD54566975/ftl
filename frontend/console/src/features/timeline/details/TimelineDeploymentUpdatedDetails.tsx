import { AttributeBadge } from '../../../components/AttributeBadge'
import type { DeploymentUpdatedEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { DeploymentCard } from '../../deployments/DeploymentCard'

export const TimelineDeploymentUpdatedDetails = ({
  deployment,
}: {
  event: Event
  deployment: DeploymentUpdatedEvent
}) => {
  return (
    <div className='p-4'>
      <DeploymentCard deploymentKey={deployment.key} />

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='Name' value={deployment.key} />
        </li>
        <li>
          <AttributeBadge name='MinReplicas' value={deployment.minReplicas.toString()} />
        </li>
        <li>
          <AttributeBadge name='PrevMinReplicas' value={deployment.prevMinReplicas.toString()} />
        </li>
      </ul>
    </div>
  )
}
