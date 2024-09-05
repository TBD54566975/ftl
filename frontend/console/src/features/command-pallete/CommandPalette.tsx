import { Combobox, ComboboxInput, ComboboxOption, ComboboxOptions, Dialog, DialogBackdrop, DialogPanel } from '@headlessui/react'
import { CellsIcon } from 'hugeicons-react'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSchema } from '../../api/schema/use-schema'
import { type PaletteItem, paletteItems } from './command-palette.utils'

type CommandPaletteProps = {
  isOpen: boolean
  onClose: () => void
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({ isOpen, onClose }) => {
  const navigate = useNavigate()
  const { data: schemaData } = useSchema()
  const [query, setQuery] = useState('')
  const [items, setItems] = useState<PaletteItem[]>([])

  useEffect(() => {
    if (schemaData) {
      const newItems = paletteItems(schemaData)
      setItems(newItems)
    }
  }, [schemaData])

  const filteredItems = query === '' ? [] : items.filter((item) => item.title.toLowerCase().includes(query.toLowerCase()))

  const handleClose = () => {
    onClose()
    setQuery('')
  }

  if (!isOpen) return null

  return (
    <Dialog className='relative z-10' open={isOpen} onClose={handleClose}>
      <DialogBackdrop
        transition
        className='fixed inset-0 bg-slate-900 bg-opacity-40 transition-opacity data-[closed]:opacity-0 data-[enter]:duration-300 data-[leave]:duration-200 data-[enter]:ease-out data-[leave]:ease-in'
      />

      <div className='fixed inset-0 z-10 w-screen overflow-y-auto p-4 sm:p-6 md:p-20'>
        <DialogPanel
          transition
          className='mx-auto max-w-xl transform rounded-xl bg-white p-2 shadow-2xl ring-1 ring-black ring-opacity-5 transition-all data-[closed]:scale-95 data-[closed]:opacity-0 data-[enter]:duration-300 data-[leave]:duration-200 data-[enter]:ease-out data-[leave]:ease-in'
        >
          <Combobox
            onChange={(item: PaletteItem) => {
              if (item) {
                navigate(item.url)
                handleClose()
              }
            }}
          >
            <ComboboxInput
              id='command-palette-search-input'
              autoFocus
              className='w-full rounded-md border-0 bg-gray-100 px-4 py-2.5 text-gray-900 focus:ring-0 sm:text-sm'
              placeholder='Search...'
              onChange={(event) => setQuery(event.target.value)}
              onBlur={handleClose}
            />

            {filteredItems.length > 0 && (
              <ComboboxOptions static className='-mb-2 max-h-72 scroll-py-2 overflow-y-auto py-2 text-sm text-gray-800'>
                {filteredItems.map((item) => (
                  <ComboboxOption
                    key={item.id}
                    value={item}
                    className='group flex cursor-default select-none rounded-md px-2 py-2 data-[focus]:bg-indigo-600 data-[focus]:text-white'
                  >
                    <div className='flex size-10 flex-none items-center justify-center rounded-lg'>
                      <item.icon className='size-5 text-gray-500 group-data-[focus]:text-white' aria-hidden='true' />
                    </div>
                    <div className='ml-2 flex-auto'>
                      <p className='text-sm font-medium text-gray-700 group-data-[focus]:text-white'>{item.title}</p>
                      <p className='mt-0.5 text-xs font-roboto-mono text-gray-500 group-data-[focus]:text-gray-300'>{item.subtitle}</p>
                    </div>
                    <div className='mr-2 flex items-center text-xs text-indigo-600 group-data-[focus]:text-indigo-300 text-right flex-none'>{item.url}</div>
                  </ComboboxOption>
                ))}
              </ComboboxOptions>
            )}

            {query !== '' && filteredItems.length === 0 && (
              <div className='px-4 py-14 text-center sm:px-14'>
                <CellsIcon className='mx-auto h-6 w-6 text-gray-400' aria-hidden='true' />
                <p className='mt-4 text-sm text-gray-900'>No items found using that search term.</p>
              </div>
            )}
          </Combobox>
        </DialogPanel>
      </div>
    </Dialog>
  )
}
