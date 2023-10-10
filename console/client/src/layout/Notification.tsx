import { Transition } from '@headlessui/react'
import { XMarkIcon } from '@heroicons/react/20/solid'
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
} from '@heroicons/react/24/outline'
import { Fragment, useContext } from 'react'
import { NotificationType, NotificationsContext } from '../providers/notifications-provider'
import { textColor } from '../utils'

export const Notification = () => {
  const { isOpen, notification, closeNotification } = useContext(NotificationsContext)

  const iconColor = () => {
    switch (notification?.type) {
      case NotificationType.Success:
        return 'text-green-400'
      case NotificationType.Error:
        return 'text-red-400'
      case NotificationType.Warning:
        return 'text-yellow-400'
      case NotificationType.Info:
        return 'text-blue-400'
      default:
        return 'text-gray-400'
    }
  }

  const icon = () => {
    switch (notification?.type) {
      case NotificationType.Success:
        return <CheckCircleIcon className={`h-6 w-6 ${iconColor()}`} aria-hidden='true' />
      case NotificationType.Error:
        return <ExclamationTriangleIcon className={`h-6 w-6 ${iconColor()}`} aria-hidden='true' />
      case NotificationType.Warning:
        return <ExclamationCircleIcon className={`h-6 w-6 ${iconColor()}`} aria-hidden='true' />
      case NotificationType.Info:
        return <InformationCircleIcon className={`h-6 w-6 ${iconColor()}`} aria-hidden='true' />
      default:
        return <></>
    }
  }

  return (
    <div
      aria-live='assertive'
      className={'pointer-events-none z-30 fixed inset-0 flex items-end px-4 py-6 sm:items-start sm:p-6'}
    >
      <div className='flex w-full flex-col items-center space-y-4 sm:items-end'>
        <Transition
          show={isOpen}
          as={Fragment}
          enter='transform ease-out duration-300 transition'
          enterFrom='translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2'
          enterTo='translate-y-0 opacity-100 sm:translate-x-0'
          leave='transition ease-in duration-100'
          leaveFrom='opacity-100'
          leaveTo='opacity-0'
        >
          <div
            className={`${textColor} pointer-events-auto w-full max-w-sm overflow-hidden rounded-lg bg-white shadow-lg ring-1 ring-black ring-opacity-5`}
          >
            <div className='p-4'>
              <div className='flex items-start'>
                <div className='flex-shrink-0'>{icon()}</div>
                <div className='ml-3 w-0 flex-1 pt-0.5'>
                  <p className='text-sm font-medium text-gray-900'>{notification?.title}</p>
                  <p className='mt-1 text-sm text-gray-500'>{notification?.message}</p>
                </div>
                <div className='ml-4 flex flex-shrink-0'>
                  <button
                    type='button'
                    className='inline-flex rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2'
                    onClick={() => {
                      closeNotification()
                    }}
                  >
                    <span className='sr-only'>Close</span>
                    <XMarkIcon className='h-5 w-5' aria-hidden='true' />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  )
}
