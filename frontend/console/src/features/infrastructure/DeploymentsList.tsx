import { AttributeBadge } from '../../components'
import { Badge } from '../../components/Badge'
import { List } from '../../components/List'
import type { StatusResponse_Deployment } from '../../protos/xyz/block/ftl/v1/ftl_pb'
import { classNames } from '../../utils'
import { deploymentTextColor } from '../deployments/deployment.utils'
import { renderValue } from './infrastructure.utils'

export const DeploymentsList = ({ deployments }: { deployments: StatusResponse_Deployment[] }) => {
  return (
    <List
      items={deployments}
      renderItem={(deployment) => (
        <div className='flex w-full'>
          <div className='flex gap-x-4 items-center w-1/2'>
            <div className='whitespace-nowrap'>
              <div className='flex gap-x-2 items-center'>
                <p>{deployment.name}</p>
                <Badge name={deployment.language} />
              </div>

              <p className={classNames(deploymentTextColor(deployment.key), 'text-sm font-semibold leading-6')}>{deployment.key}</p>
            </div>
          </div>
          <div className='flex gap-x-4 items-center w-1/2 justify-end'>
            <div className='flex flex-wrap gap-1'>
              <AttributeBadge key='replicas' name='replicas' value={deployment.replicas.toString()} />
              <AttributeBadge key='min_replicas' name='min_replicas' value={deployment.minReplicas.toString()} />
              {Object.entries(deployment.labels?.fields || {}).map(([key, value]) => (
                <AttributeBadge key={key} name={key} value={renderValue(value)} />
              ))}
            </div>
          </div>
        </div>
      )}
    />
  )
}
