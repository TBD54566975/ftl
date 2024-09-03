import { AttributeBadge } from '../../components'
import { List } from '../../components/List'
import type { StatusResponse_Route } from '../../protos/xyz/block/ftl/v1/ftl_pb'

export const RoutesList = ({ routes }: { routes: StatusResponse_Route[] }) => {
  return (
    <List
      items={routes}
      renderItem={(route) => (
        <div className='flex w-full'>
          <div className='flex gap-x-4 items-center w-1/2'>
            <div className='whitespace-nowrap'>
              <div className='flex gap-x-2 items-center'>{route.module}</div>
              <p className='mt-1 flex text-xs leading-5 text-gray-500 dark:text-gray-400 font-roboto-mono'>{route.endpoint}</p>
            </div>
          </div>
          <div className='flex gap-x-4 items-center w-1/2 justify-end'>
            <div className='flex flex-wrap gap-1 justify-end'>
              <AttributeBadge name='deployment' value={route.deployment} />
              <AttributeBadge name='runner' value={route.runner} />
            </div>
          </div>
        </div>
      )}
    />
  )
}
