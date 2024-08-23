import { Disclosure, DisclosureButton, DisclosurePanel } from '@headlessui/react'
import { ChevronRightIcon } from '@heroicons/react/24/outline'
import { classNames } from '../../utils'
import type { ModuleTreeItem } from './module.utils'

export const ModulesTree = ({ modules }: { modules: ModuleTreeItem[] }) => {
  return (
    <div className='flex grow flex-col h-full gap-y-5 overflow-y-auto border-r border-gray-200 dark:border-gray-600 py-2 px-6'>
      <nav className='flex flex-1 flex-col'>
        <ul className='flex flex-1 flex-col gap-y-7'>
          <li>
            <ul className='-mx-2'>
              {modules.map((item) => (
                <li key={item.name} id={`module-tree-item-${item.name}`}>
                  {!item.children ? (
                    <a
                      href={item.href}
                      className={classNames(
                        item.current ? '' : 'hover:bg-gray-50 hover:dark:bg-gray-700',
                        'group flex gap-x-3 rounded-md px-2 text-sm font-semibold leading-6',
                      )}
                    >
                      <item.icon aria-hidden='true' className='size-3 shrink-0' />
                      {item.name}
                    </a>
                  ) : (
                    <Disclosure as='div' defaultOpen={item.expanded}>
                      <DisclosureButton
                        className={classNames(
                          item.current ? '' : 'hover:bg-gray-50 hover:dark:bg-gray-700',
                          'group flex w-full items-center gap-x-2 rounded-md px-2 text-left text-sm font-semibold leading-6',
                        )}
                      >
                        <item.icon aria-hidden='true' className='size-4 shrink-0 ' />
                        {item.name}
                        <ChevronRightIcon aria-hidden='true' className='ml-auto h-5 w-5 shrink-0 group-data-[open]:rotate-90 group-data-[open]:text-gray-500' />
                      </DisclosureButton>
                      <DisclosurePanel as='ul' className='px-2'>
                        {item.children.map((subItem) => (
                          <li key={subItem.name}>
                            <DisclosureButton
                              as='a'
                              href={subItem.href}
                              className={classNames(
                                subItem.current ? '' : 'hover:bg-gray-50 hover:dark:bg-gray-700',
                                'group flex items-center gap-x-2 rounded-md pl-4 pr-2 text-sm leading-6',
                              )}
                            >
                              <subItem.icon aria-hidden='true' className='size-4 shrink-0' />
                              {subItem.name}
                            </DisclosureButton>
                          </li>
                        ))}
                      </DisclosurePanel>
                    </Disclosure>
                  )}
                </li>
              ))}
            </ul>
          </li>
        </ul>
      </nav>
    </div>
  )
}
