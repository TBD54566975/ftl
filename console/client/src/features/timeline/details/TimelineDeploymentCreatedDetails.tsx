import { Timestamp } from '@bufbuild/protobuf'
import React from 'react'
import { AttributeBadge } from '../../../components/AttributeBadge'
import { CloseButton } from '../../../components/CloseButton'
import { DeploymentCreatedEvent, Event } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { SidePanelContext } from '../../../providers/side-panel-provider'
import { TimelineTimestamp } from './TimelineTimestamp'

interface Props {
  event: Event
  deployment: DeploymentCreatedEvent
}

export const TimelineDeploymentCreatedDetails = ({ event, deployment }: Props) => {
  const { closePanel } = React.useContext(SidePanelContext)
  return (
    <>
      <div className={`bg-green-500 dark:bg-green-300 h-2 w-full`}></div>
      <div className='p-4'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center space-x-2'>
            <span
              className={
                'text-green-500 bg-green-400/30 dark:text-green-300 dark:bg-green-700/10 inline-flex items-center rounded-md px-2 py-1 text-xs font-medium'
              }
            >
              Deployment Created
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
