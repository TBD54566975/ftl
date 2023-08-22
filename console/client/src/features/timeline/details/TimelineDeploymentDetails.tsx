import { Deployment, StreamTimelineResponse } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { classNames } from '../../../utils/react.utils'
import { textColor } from '../../../utils/style.utils'
import { TimelineTimestamp } from './TimelineTimestamp'

type Props = {
  entry: StreamTimelineResponse
  deployment: Deployment
}

export const deploymentTypeText = {
  0: 'Created',
  1: 'Updated',
  2: 'Replaced',
}

export const deploymentTypeBadge = {
  0: 'text-green-600 bg-green-400/30 dark:text-green-300 dark:bg-green-700/10',
  1: 'text-blue-350 bg-blue-300/30 dark:text-blue-300 dark:bg-blue-700/30',
  2: 'text-indigo-600 bg-indigo-400/30 dark:text-indigo-300 dark:bg-indigo-700/10',
}

export const TimelineDeploymentDetails: React.FC<Props> = ({ entry, deployment }) => {
  return (
    <>
      <div>
        <TimelineTimestamp entry={entry} />
      </div>

      <div className='pt-4'>
        <span className={classNames(deploymentTypeBadge[deployment.eventType], 'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium text-gray-600')}>
          {deploymentTypeText[deployment.eventType]}
        </span>
      </div>

      <div className='pt-2 text-gray-500 dark:text-gray-400'>
        <div className='flex pt-2 justify-between'>
          <dt>Name</dt>
          <dd className={`${textColor}`}>{deployment.name}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Module</dt>
          <dd className={`${textColor}`}>{deployment.moduleName}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Language</dt>
          <dd className={`${textColor}`}>{deployment.language}</dd>
        </div>
        <div className='flex pt-2 justify-between'>
          <dt>Min replicas</dt>
          <dd className={`${textColor}`}>{deployment.minReplicas}</dd>
        </div>
        {deployment.replaced && (
          <div className='flex pt-2 justify-between'>
            <dt>Replaced</dt>
            <dd className={`${textColor}`}>{deployment.replaced}</dd>
          </div>
        )}
      </div>
    </>
  )
}
