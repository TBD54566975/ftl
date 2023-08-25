import { Timestamp } from '@bufbuild/protobuf'
import { RocketLaunchIcon } from '@heroicons/react/24/solid'
import { Deployment, DeploymentEventType } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { formatTimestamp } from '../../utils/date.utils'
import { panelColor, textColor } from '../../utils/style.utils'
import { classNames } from '../../utils/react.utils'

type Props = {
  deployment: Deployment
  timestamp?: Timestamp
  selected?: boolean
}

function deploymentType(type: DeploymentEventType) {
  switch (type) {
    case DeploymentEventType.DEPLOYMENT_CREATED: return 'Created'
    case DeploymentEventType.DEPLOYMENT_UPDATED: return 'Updated'
    case DeploymentEventType.DEPLOYMENT_REPLACED: return 'Replaced'
    default: return 'Unknown'
  }
}

export const TimelineDeployment: React.FC<Props> = ({ deployment, timestamp, selected }) => {
  return (
    <>
      <div className={`relative flex h-6 w-6 flex-none items-center justify-center ${panelColor}`}>
        <RocketLaunchIcon className='h-6 w-6 text-indigo-500'
          aria-hidden='true'
        />
      </div>
      <div className={classNames(`relative flex gap-x-4 flex-auto w-full max-w-full p-1.5 `, selected && 'bg-indigo-600 rounded-md')}>
        <p className={`flex-auto text-xs leading-5 ${selected ? 'text-white' : textColor}`}>
          {deploymentType(deployment.eventType)}
          {' '}deployment{' '}
          <span>{deployment.name}</span>
          {' '}for language{' '}
          <span>{deployment.language}</span>.
        </p>
        <time
          dateTime={formatTimestamp(timestamp)}
          className={`flex-none text-xs leading-5 ${selected ? 'text-gray=50': 'text-gray-500'}`}
        >
          {formatTimestamp(timestamp)}
        </time>
      </div>
    </>
  )
}
