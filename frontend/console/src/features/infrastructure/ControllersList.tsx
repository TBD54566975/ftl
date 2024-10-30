import { Badge } from '../../components/Badge'
import { List } from '../../components/List'
import type { StatusResponse_Controller } from '../../protos/xyz/block/ftl/v1/controller_pb'

export const ControllersList = ({ controllers }: { controllers: StatusResponse_Controller[] }) => {
  return (
    <List
      items={controllers}
      renderItem={(controller) => (
        <>
          <div className='flex min-w-0 gap-x-4'>
            <div className='min-w-0 flex-auto'>
              <p className='text-sm font-semibold leading-6'>
                <span className='absolute inset-x-0 -top-px bottom-0' />
                {controller.key}
              </p>
              <p className='mt-1 flex text-xs leading-5 text-gray-500 dark:text-gray-400 font-roboto-mono'>{controller.endpoint}</p>
            </div>
          </div>
          <div className='flex shrink-0 items-center gap-x-4'>
            <div className='hidden sm:flex sm:flex-col sm:items-end'>
              <Badge name={controller.version} />
            </div>
          </div>
        </>
      )}
    />
  )
}
