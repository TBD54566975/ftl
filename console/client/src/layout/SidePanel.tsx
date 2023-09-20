import { Transition } from '@headlessui/react'
import { Fragment, useContext } from 'react'
import { SidePanelContext } from '../providers/side-panel-provider'
import { sidePanelColor } from '../utils/style.utils'

export const SidePanel = () => {
  const { isOpen, component } = useContext(SidePanelContext)

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
        leaveTo='translate-x-full'
      >
        <div
          className={`fixed right-0 w-1/3 h-full ${sidePanelColor} dark:bg-slate-800 dark:shadow-black-600 shadow-2xl`}
        >
          <div className='overflow-y-auto' style={{ maxHeight: 'calc(100vh - 2rem)' }}>
            {component}
          </div>
        </div>
      </Transition>
    </div>
  )
}
