import { Timestamp } from '@bufbuild/protobuf'
import React from 'react'
import { AttributeBadge } from '../../../components/AttributeBadge'
import { CloseButton } from '../../../components/CloseButton'
import { Deployment, DeploymentEventType, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../../providers/side-panel-provider'
import { classNames } from '../../../utils/react.utils'
import { TimelineTimestamp } from './TimelineTimestamp'

interface Props {
  event: Event
  deployment: Deployment
}

export const deploymentTypeText: { [key in DeploymentEventType]: string } = {
  [DeploymentEventType.DEPLOYMENT_UNKNOWN]: 'Unknown',
  [DeploymentEventType.DEPLOYMENT_CREATED]: 'Created',
  [DeploymentEventType.DEPLOYMENT_UPDATED]: 'Updated',
  [DeploymentEventType.DEPLOYMENT_REPLACED]: 'Replaced',
}

export const deploymentTypeBarColor: { [key in DeploymentEventType]: string } = {
  [DeploymentEventType.DEPLOYMENT_UNKNOWN]: '',
  [DeploymentEventType.DEPLOYMENT_CREATED]: 'bg-green-500 dark:bg-green-300',
  [DeploymentEventType.DEPLOYMENT_UPDATED]: 'bg-blue-500 dark:bg-blue-300',
  [DeploymentEventType.DEPLOYMENT_REPLACED]: 'bg-indigo-600 dark:bg-indigo-300',
}

export const deploymentTypeBadge: { [key in DeploymentEventType]: string } = {
  [DeploymentEventType.DEPLOYMENT_UNKNOWN]: '',
  [DeploymentEventType.DEPLOYMENT_CREATED]: 'text-green-500 bg-green-400/30 dark:text-green-300 dark:bg-green-700/10',
  [DeploymentEventType.DEPLOYMENT_UPDATED]: 'text-blue-500 bg-blue-300/30 dark:text-blue-300 dark:bg-blue-700/30',
  [DeploymentEventType.DEPLOYMENT_REPLACED]:
    'text-indigo-600 bg-indigo-400/30 dark:text-indigo-300 dark:bg-indigo-700/50',
}

export const TimelineDeploymentDetails = ({ event, deployment }: Props) => {
  const { closePanel } = React.useContext(SidePanelContext)
  return (
    <>
      <div className={`${deploymentTypeBarColor[deployment.eventType]} h-2 w-full`}></div>
      <div className='p-4'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center space-x-2'>
            <span
              className={classNames(
                deploymentTypeBadge[deployment.eventType],
                'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium text-gray-600',
              )}
            >
              {deploymentTypeText[deployment.eventType]}
            </span>
            <TimelineTimestamp timestamp={event.timeStamp ?? new Timestamp()} />
          </div>
          <CloseButton onClick={closePanel} />
        </div>

        <ul className='pt-4 space-y-2'>
          <li>
            <AttributeBadge name='Name' value={deployment.name} />
          </li>
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
    </>
  )
}
