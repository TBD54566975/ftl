import { AttributeBadge } from '../../../components/AttributeBadge'
import type { DeploymentUpdatedEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
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
          <AttributeBadge name='name' value={deployment.key} />
        </li>
        <li>
          <AttributeBadge name='min_replicas' value={deployment.minReplicas.toString()} />
        </li>
        <li>
          <AttributeBadge name='prev_min_replicas' value={deployment.prevMinReplicas.toString()} />
        </li>
      </ul>
    </div>
  )
}
