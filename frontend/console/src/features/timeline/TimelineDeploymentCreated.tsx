import type { DeploymentCreatedEvent } from '../../protos/xyz/block/ftl/console/v1/console_pb'

export const TimelineDeploymentCreated = ({ deployment }: { deployment: DeploymentCreatedEvent }) => {
  const title = `Created deployment ${deployment.key} for language ${deployment.language}`
  return (
    <span title={title}>
      Created deployment <span className='text-indigo-500 dark:text-indigo-300'>{deployment.key}</span> for language{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>{deployment.language}</span>
    </span>
  )
}
