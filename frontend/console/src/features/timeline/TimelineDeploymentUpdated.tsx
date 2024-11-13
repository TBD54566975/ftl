import type { DeploymentUpdatedEvent } from '../../protos/xyz/block/ftl/console/v1/console_pb'

export const TimelineDeploymentUpdated = ({ deployment }: { deployment: DeploymentUpdatedEvent }) => {
  const title = `Updated deployment ${deployment.key} min replicas to ${deployment.minReplicas} (previously ${deployment.prevMinReplicas})`
  return (
    <span title={title}>
      Updated deployment <span className='text-indigo-500 dark:text-indigo-300'>{deployment.key}</span> min replicas to{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>{deployment.minReplicas}</span> (previously{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>{deployment.prevMinReplicas}</span>)
    </span>
  )
}
