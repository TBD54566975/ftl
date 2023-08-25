import {Transition} from '@headlessui/react'
import {XMarkIcon} from '@heroicons/react/24/outline'
import {Fragment, useContext} from 'react'
import {SidePanelContext} from '../providers/side-panel-provider'
import {
  headerColor,
  headerTextColor,
  sidePanelColor,
} from '../utils/style.utils'

export function SidePanel() {
  const {isOpen, closePanel, component} = useContext(SidePanelContext)

  return (
    <div className='relative z-20'>
      <Transition
        show={isOpen}
        as={Fragment}
        enter='transform transition ease-in-out duration-300'
        enterFrom='translate-x-full'
        enterTo='translate-x-0'
        leave='transform transition ease-in-out duration-300'
        leaveFrom='translate-x-0'
        leaveTo='translate-x-full'>
        <div
          className={`fixed right-0 w-1/3 h-full mt-2 ${sidePanelColor} shadow-xl`}>
          <div
            className={`flex items-center justify-between pl-4 pr-2 py-2 rounded-tl ${headerTextColor} ${headerColor}`}>
            Event Details
            <button
              onClick={() => closePanel()}
              className='hover:bg-indigo-800 dark:hover:bg-indigo-500'>
              <XMarkIcon className={`h-5 w-5`} />
            </button>
          </div>
          <div
            className='p-4 overflow-y-auto'
            style={{maxHeight: 'calc(100vh - 6rem)'}}>
            {component}
          </div>
        </div>
      </Transition>
    </div>
  )
}
