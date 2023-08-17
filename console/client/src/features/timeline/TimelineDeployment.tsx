import { Timestamp } from '@bufbuild/protobuf'
import { RocketLaunchIcon } from '@heroicons/react/24/solid'
import { Deployment, DeploymentEventType } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatTimestamp } from '../../utils/date.utils'

type Props = {
  deployment: Deployment
  timestamp?: Timestamp
}

function deploymentType(type: DeploymentEventType) {
  switch (type) {
    case DeploymentEventType.DEPLOYMENT_CREATED: return 'Created'
    case DeploymentEventType.DEPLOYMENT_REPLACED: return 'Replaced'
    default: return 'Unknown'
  }
}

export const TimelineDeployment: React.FC<Props> = ({ deployment, timestamp }) => {
  return (
    <>
      <div className='relative flex h-6 w-6 flex-none items-center justify-center bg-white dark:bg-slate-800'>
        <RocketLaunchIcon className='h-6 w-6 text-indigo-500'
          aria-hidden='true'
        />
      </div>
      <p className='flex-auto py-0.5 text-xs leading-5 text-gray-500 dark:text-gray-400'>
        {deploymentType(deployment.eventType)} deployment <span className={`font-medium text-gray-900 dark:text-white`}>{deployment.name}</span>{' '}
                  for language{' '}
        <span className='font-medium text-gray-900 dark:text-white'>{deployment.language}</span>.
      </p>
      <time
        dateTime={formatTimestamp(timestamp)}
        className='flex-none py-0.5 text-xs leading-5 text-gray-500'
      >
        {formatTimestamp(timestamp)}
      </time>
    </>
  )
}
