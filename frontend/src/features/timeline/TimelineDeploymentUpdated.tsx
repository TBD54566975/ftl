import type { DeploymentUpdatedEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'

export const TimelineDeploymentUpdated = ({ deployment }: { deployment: DeploymentUpdatedEvent }) => {
  return (
    <>
      Updated deployment <span className='text-indigo-500 dark:text-indigo-300'>{deployment.key}</span> min replicas to{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>{deployment.minReplicas}</span> (previously{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>{deployment.prevMinReplicas}</span>)
    </>
  )
}
