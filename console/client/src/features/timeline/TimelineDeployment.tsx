import {
  Deployment,
  DeploymentEventType,
} from '../../protos/xyz/block/ftl/v1/console/console_pb'

type Props = {
  deployment: Deployment
}

function deploymentType(type: DeploymentEventType) {
  switch (type) {
    case DeploymentEventType.DEPLOYMENT_CREATED:
      return 'Created'
    case DeploymentEventType.DEPLOYMENT_UPDATED:
      return 'Updated'
    case DeploymentEventType.DEPLOYMENT_REPLACED:
      return 'Replaced'
    default:
      return 'Unknown'
  }
}

export const TimelineDeployment: React.FC<Props> = ({deployment}) => {
  return (
    <>
      <span>{deploymentType(deployment.eventType)}</span> deployment{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>
        {deployment.name}
      </span>{' '}
      for language{' '}
      <span className='text-indigo-500 dark:text-indigo-300'>
        {deployment.language}
      </span>
    </>
  )
}
