import { DeploymentCreatedEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'

interface Props {
  deployment: DeploymentCreatedEvent
}

export const TimelineDeploymentCreated = ({ deployment }: Props) => {
  return (
    <>
      Created deployment <span className='text-indigo-500 dark:text-indigo-300'>{deployment.name}</span> for language{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>{deployment.language}</span>
    </>
  )
}
