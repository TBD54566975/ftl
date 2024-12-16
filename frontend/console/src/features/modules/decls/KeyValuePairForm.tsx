import { Delete03Icon } from 'hugeicons-react'
import { useEffect, useRef } from 'react'
import { Checkbox } from '../../../components/Checkbox'

interface KeyValuePair {
  id: string
  enabled: boolean
  key: string
  value: string
}

interface KeyValuePairFormProps {
  keyValuePairs: KeyValuePair[]
  onChange: (pairs: KeyValuePair[]) => void
}

export const KeyValuePairForm = ({ keyValuePairs, onChange }: KeyValuePairFormProps) => {
  const inputRefs = useRef<{ [key: string]: HTMLInputElement | null }>({})

  useEffect(() => {
    if (keyValuePairs.length === 0) {
      onChange([{ id: crypto.randomUUID(), enabled: false, key: '', value: '' }])
    }
  }, [])

  const updatePair = (id: string, updates: Partial<KeyValuePair>) => {
    const updatedPairs = keyValuePairs.map((p) => {
      if (p.id === id) {
        const updatedPair = { ...p, ...updates }
        if ('key' in updates) {
          updatedPair.enabled = (updates.key?.length ?? 0) > 0
        }
        return updatedPair
      }
      return p
    })

    const filteredPairs = updatedPairs.filter((p, index) => index === updatedPairs.length - 1 || !(p.key === '' && p.value === ''))

    const lastPair = filteredPairs[filteredPairs.length - 1]
    if (lastPair?.key || lastPair?.value) {
      const newId = crypto.randomUUID()
      filteredPairs.push({ id: newId, enabled: false, key: '', value: '' })
    }

    onChange(filteredPairs)

    // Focus the key input of the last empty row after deletion
    if ('key' in updates && updates.key === '') {
      setTimeout(() => {
        const lastPairId = filteredPairs[filteredPairs.length - 1].id
        inputRefs.current[lastPairId]?.focus()
      }, 0)
    }
  }

  return (
    <div className='p-4'>
      <div className='space-y-2'>
        {keyValuePairs.map((pair, index) => (
          <div key={pair.id} className='flex items-center gap-2'>
            <Checkbox checked={pair.enabled} onChange={(e) => updatePair(pair.id, { enabled: e.target.checked })} />
            <input
              ref={(el) => {
                inputRefs.current[pair.id] = el
              }}
              type='text'
              value={pair.key}
              onChange={(e) => updatePair(pair.id, { key: e.target.value })}
              placeholder='Key'
              className='flex-1 block rounded-md border-0 py-1.5 text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 dark:bg-gray-800 sm:text-sm sm:leading-6'
            />
            <input
              type='text'
              value={pair.value}
              onChange={(e) => updatePair(pair.id, { value: e.target.value })}
              placeholder='Value'
              className='flex-1 block rounded-md border-0 py-1.5 text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 dark:bg-gray-800 sm:text-sm sm:leading-6'
            />
            <div className='w-8 flex-shrink-0'>
              {(pair.key || pair.value || keyValuePairs.length > 1) && index !== keyValuePairs.length - 1 && (
                <button type='button' onClick={() => updatePair(pair.id, { key: '', value: '' })} className='p-2 text-gray-500 hover:text-gray-500'>
                  <Delete03Icon className='h-5 w-5' />
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
