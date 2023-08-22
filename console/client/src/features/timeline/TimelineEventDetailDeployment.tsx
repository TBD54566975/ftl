import { Deployment } from '../../protos/xyz/block/ftl/v1/console/console_pb'


type Props = {
  deployment: Deployment
}

export const TimelineEventDetailDeployment: React.FC<Props> = ({ deployment }) => {
  return (
    <>
      <div>
        {deployment.eventType}
      </div>
    </>
  )
}
