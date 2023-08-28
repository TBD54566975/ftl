import React from 'react'
import {Transition} from '@headlessui/react'
import {Portal} from '@headlessui/react'

export const Notification: React.FC<{
  title: string
  message: string
  icon?: React.ReactNode
  color: string
}> = ({title, message, icon, color}) => {
  const [active, setActive] = React.useState(false)
  React.useEffect(() => {
    const timer = setTimeout(() => {
      setActive(true)
    }, 50)

    return () => clearTimeout(timer)
  }, [])
  return (
    <Portal>
      <Transition
        show={active}
        enter='transition-opacity duration-150'
        enterFrom='opacity-0'
        enterTo='opacity-150'
        leave='transition-opacity duration-150'
        leaveFrom='opacity-150'
        leaveTo='opacity-0'
      >
        <div
          className={`${color} z-10 bg-opacity-5 fixed bottom-0.5 left-1/2 flex items-center p-4 mb-4 text-sm  rounded-lg bg-gray-800 gap-1`}
          role='alert'
          style={
            {
              transform: `translateX(-50%)`,
              width: `calc(100% - 2rem)`,
              '--tw-bg-opacity': '0.85',
            } as React.CSSProperties
          }
        >
          {icon}
          <span className='font-medium'>{title}</span>
          <span>{message}</span>
        </div>
      </Transition>
    </Portal>
  )
}
