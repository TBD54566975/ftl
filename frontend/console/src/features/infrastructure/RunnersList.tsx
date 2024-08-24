import { AttributeBadge } from '../../components'
import { List } from '../../components/List'
import { StatusIndicator } from '../../components/StatusIndicator'
import { RunnerState, type StatusResponse_Runner } from '../../protos/xyz/block/ftl/v1/ftl_pb'
import { classNames } from '../../utils'
import { deploymentTextColor } from '../deployments/deployment.utils'
import { renderValue } from './infrastructure.utils'

export const RunnersList = ({ runners }: { runners: StatusResponse_Runner[] }) => {
  return (
    <List
      items={runners}
      renderItem={(runner) => (
        <>
          <div className='flex gap-x-4 items-center'>
            <div className='whitespace-nowrap'>
              <p className='text-sm font-semibold leading-6'>{runner.key}</p>
              <p className='mt-1 flex text-xs leading-5 text-gray-500 dark:text-gray-400 font-roboto-mono'>{runner.endpoint}</p>
              <div className='mt-1 flex gap-x-2 items-center'>
                {status(runner.state)}
                {runner.deployment && <p className={classNames(deploymentTextColor(runner.deployment), 'text-xs')}>{runner.deployment}</p>}
              </div>
            </div>
          </div>
          <div className='flex gap-x-4 items-center w-1/2'>
            <div className='flex flex-wrap gap-1 justify-end'>
              {Object.entries(runner.labels?.fields || {}).map(([key, value]) => (
                <AttributeBadge key={key} name={key} value={renderValue(value)} />
              ))}
            </div>
          </div>
        </>
      )}
    />
  )
}

const status = (state: RunnerState) => {
  switch (state) {
    case RunnerState.RUNNER_ASSIGNED:
      return <StatusIndicator state='success' text='Assigned' />
    case RunnerState.RUNNER_DEAD:
      return <StatusIndicator state='error' text='Dead' />
    case RunnerState.RUNNER_NEW:
      return <StatusIndicator state='new' text='New' />
  }
}
