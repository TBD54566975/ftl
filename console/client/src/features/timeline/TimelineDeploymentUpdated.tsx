import { DeploymentUpdatedEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'

interface Props {
  deployment: DeploymentUpdatedEvent
}

export const TimelineDeploymentUpdated = ({ deployment }: Props) => {
  return (
    <>
      Updated deployment <span className='text-indigo-500 dark:text-indigo-300'>{deployment.name}</span> min replicas to{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>{deployment.minReplicas}</span> (previously{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>{deployment.prevMinReplicas}</span>)
    </>
  )
}
