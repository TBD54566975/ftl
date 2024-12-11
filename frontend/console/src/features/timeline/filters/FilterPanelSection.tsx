import { Disclosure } from '@headlessui/react'
import { ArrowUp01Icon } from 'hugeicons-react'
import { textColor } from '../../../utils'

export const FilterPanelSection = ({
  title,
  children,
  defaultOpen = true,
}: {
  title: string
  children: React.ReactNode
  defaultOpen?: boolean
}) => {
  return (
    <Disclosure defaultOpen={defaultOpen}>
      {({ open }) => (
        <div>
          <Disclosure.Button
            className={`flex w-full justify-between rounded-md bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 py-1 px-2 text-left text-sm font-medium ${textColor} focus:outline-none focus-visible:ring focus-visible:ring-gray-500 focus-visible:ring-opacity-75`}
          >
            <span>{title}</span>
            <ArrowUp01Icon className={`${open ? 'rotate-180 transform' : ''} h-5 w-5 text-gray-500`} />
          </Disclosure.Button>
          <Disclosure.Panel className={`px-2 py-2 text-sm ${textColor}`}>
            <fieldset>
              <legend className='sr-only'>{title}</legend>
              <div className='space-y-0.5'>{children}</div>
            </fieldset>
          </Disclosure.Panel>
        </div>
      )}
    </Disclosure>
  )
}
