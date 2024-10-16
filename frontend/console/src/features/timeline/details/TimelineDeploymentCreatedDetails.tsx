import { AttributeBadge } from '../../../components/AttributeBadge'
import type { DeploymentCreatedEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { DeploymentCard } from '../../deployments/DeploymentCard'

export const TimelineDeploymentCreatedDetails = ({
  deployment,
}: {
  event: Event
  deployment: DeploymentCreatedEvent
}) => {
  return (
    <div className='p-4'>
      <DeploymentCard deploymentKey={deployment.key} />

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='Module' value={deployment.moduleName} />
        </li>
        <li>
          <AttributeBadge name='Language' value={deployment.language} />
        </li>
        <li>
          <AttributeBadge name='MinReplicas' value={deployment.minReplicas.toString()} />
        </li>
        {deployment.replaced && (
          <li>
            <AttributeBadge name='Replaced' value={deployment.replaced} />
          </li>
        )}
      </ul>
    </div>
  )
}
