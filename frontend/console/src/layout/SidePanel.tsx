import { Transition } from '@headlessui/react'
import { Fragment, useContext } from 'react'
import { SidePanelContext } from '../providers/side-panel-provider'
import { bgColor, textColor } from '../utils'
import { sidePanelColor } from '../utils/style.utils'

export const SidePanel = () => {
  const { isOpen, header, component } = useContext(SidePanelContext)

  return (
    <div className={`absolute z-20 top-0 ${bgColor} ${textColor}`}>
      <Transition
        show={isOpen}
        as={Fragment}
        enter='transform transition ease-in-out duration-300'
        enterFrom='translate-x-full'
        enterTo='translate-x-0'
        leave='transform transition ease-in-out duration-300'
        leaveFrom='translate-x-0'
        leaveTo='translate-x-full'
      >
        <div className={`fixed inset-y-0 right-0 sm:w-1/3 w-3/4 ${sidePanelColor} dark:bg-slate-800 dark:shadow-black-600 shadow-2xl flex flex-col`}>
          {header && <div className='flex-shrink-0'>{header}</div>}
          <div className='overflow-y-auto h-full'>{component}</div>
        </div>
      </Transition>
    </div>
  )
}
