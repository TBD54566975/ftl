import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from '@headlessui/react'
import { ArrowDown01Icon, CheckmarkSquare02Icon, SquareIcon } from 'hugeicons-react'
import { Divider } from './Divider'

export interface MultiselectOpt {
  key: string
  displayName: string
}

export function sortMultiselectOpts(o: MultiselectOpt[]) {
  return o.sort((a: MultiselectOpt, b: MultiselectOpt) => (a.key < b.key ? -1 : 1))
}

const getSelectionText = (selectedOpts: MultiselectOpt[], allOpts: MultiselectOpt[]): string => {
  if (selectedOpts.length === 0) {
    return 'Select types...'
  }
  if (selectedOpts.length === allOpts.length) {
    return 'Filter types...'
  }
  return selectedOpts.map((o) => o.displayName).join(', ')
}

export const Multiselect = ({
  allOpts,
  selectedOpts,
  onChange,
}: { allOpts: MultiselectOpt[]; selectedOpts: MultiselectOpt[]; onChange: (types: MultiselectOpt[]) => void }) => {
  sortMultiselectOpts(selectedOpts)
  return (
    <div className='w-full'>
      <Listbox multiple value={selectedOpts} onChange={onChange}>
        <div className='relative w-[calc(100%-0.75rem)]'>
          <ListboxButton className='w-full m-1.5 py-1 px-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm text-gray-600 dark:text-gray-300 bg-white dark:bg-gray-900 hover:text-gray-800 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-700'>
            <span className='block truncate w-[calc(100%-30px)] h-5 text-left'>{getSelectionText(selectedOpts, allOpts)}</span>
            <span className='pointer-events-none absolute inset-y-0 right-0 flex items-center pr-1'>
              <ArrowDown01Icon className='w-5 text-gray-400 dark:text-gray-500' />
            </span>
          </ListboxButton>
        </div>
        <ListboxOptions
          anchor='bottom'
          transition
          className='w-[var(--button-width)] min-w-48 mt-1 pt-1 rounded-md border dark:border-white/5 bg-white dark:bg-gray-800 transition duration-100 ease-in truncate drop-shadow-lg z-20'
        >
          {allOpts.map((o) => (
            <ListboxOption
              className='cursor-pointer py-1 px-2 group flex items-center gap-2 select-none text-sm text-gray-800 dark:text-gray-200 hover:bg-gray-200 hover:dark:bg-gray-700'
              key={o.key}
              value={o}
            >
              {({ selected }) => (
                <div className='flex items-center gap-2'>
                  {selected ? <CheckmarkSquare02Icon className='size-5' /> : <SquareIcon className='size-5' />}
                  {o.displayName}
                </div>
              )}
            </ListboxOption>
          ))}

          <div className='w-full text-center text-xs mt-2'>
            <Divider />
            <div className='flex'>
              <div
                className='flex-1 p-2 hover:bg-gray-200 dark:hover:bg-gray-700 text-indigo-600 dark:text-indigo-400 cursor-pointer text-center'
                onClick={() => onChange(allOpts)}
              >
                Select all
              </div>
              <Divider vertical />
              <div
                className='flex-1 p-2 hover:bg-gray-200 dark:hover:bg-gray-700 text-indigo-600 dark:text-indigo-400 cursor-pointer text-center'
                onClick={() => onChange([])}
              >
                Deselect all
              </div>
            </div>
          </div>
        </ListboxOptions>
      </Listbox>
    </div>
  )
}
