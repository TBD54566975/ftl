import { Listbox } from '@headlessui/react'
import { ArrowDown01Icon, Tick01Icon } from 'hugeicons-react'

export interface MultiselectOpt {
  key: string
  displayName: string
}

export function sortMultiselectOpts(o: MultiselectOpt[]) {
  return o.sort((a: MultiselectOpt, b: MultiselectOpt) => (a.key < b.key ? -1 : 1))
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
        <div className='relative w-[calc(100%-0.75rem)] items-center'>
          <Listbox.Button className='w-full m-1.5 py-1 px-2 text-sm border-gray-200 dark:border-white/5 rounded-md text-sm text-gray-500 hover:text-gray-900 hover:dark:text-gray-200 bg-gray-300 dark:bg-gray-800'>
            <span className='block truncate w-[calc(100%-30px)] h-5'>{selectedOpts.map((o) => o.displayName).join(', ')}</span>
            <span className='pointer-events-none absolute inset-y-0 right-0 flex items-center pr-1'>
              <ArrowDown01Icon className='w-5' />
            </span>
          </Listbox.Button>
        </div>
        <Listbox.Options
          anchor='bottom'
          transition
          className='w-[var(--button-width)] min-w-48  rounded-xl mt-1 border dark:border-white/5 bg-white/90 dark:bg-gray-700/90 transition duration-100 ease-in truncate'
        >
          {allOpts.map((o) => (
            <Listbox.Option
              className='cursor-pointer py-1.5 px-2 group flex items-center gap-2 select-none text-sm dark:text-white hover:bg-gray-200 hover:dark:bg-gray-800'
              key={o.key}
              value={o}
            >
              <div className='w-4'>
                <Tick01Icon className='size-4 invisible group-data-[selected]:visible' />
              </div>
              {o.displayName}
            </Listbox.Option>
          ))}
          <div className='w-full mt-2 text-center text-sm dark:text-white border-t dark:border-white/5 bg-white dark:bg-gray-700'>
            <div
              className='w-1/2 p-2 inline-block hover:bg-gray-200 hover:dark:bg-white/10 cursor-pointer text-clip border-r dark:border-white/5'
              onClick={() => onChange(allOpts)}
            >
              Select All
            </div>
            <div className='w-1/2 p-2 inline-block hover:bg-gray-200 hover:dark:bg-white/10 cursor-pointer text-clip' onClick={() => onChange([])}>
              Deselect All
            </div>
          </div>
        </Listbox.Options>
      </Listbox>
    </div>
  )
}
