import { Listbox } from '@headlessui/react'
import { EyeIcon } from '@heroicons/react/20/solid'
import React, { useState } from 'react'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { classNames } from '../../utils'
import { backgrounds, borders, colors } from './components'
import { VerbId, ZoomCallbacks } from './modules.constants'
import { getNames } from './modules.utils'

interface DeploymentVerbs {
  deploymentName: string
  id: string
  verbs: Set<VerbId>
  queriedVerbs: Set<VerbId>
}

const ModulesOption = ({
  id,
  zoomCallbacks,
  verbs,
  deploymentName,
}: {
  id: string
  verbs: VerbId[]
  deploymentName: string
  zoomCallbacks?: ZoomCallbacks
}) => {
  return (
    <li className={`flex flex-wrap gap-1 ${backgrounds.level1} ${borders.level1}`}>
      <div className={`w-1 bg-${colors.deployment}`}></div>
      <div className='flex flex-wrap flex-1 justify-between items-center p-2'>
        <span className='text-black dark:text-white text-base px-1 '>{id}</span>
        <button onClick={() => zoomCallbacks?.to(id)}>
          <span className='sr-only'>Zoom to ${id} module</span>
          <EyeIcon className='cursor-pointer w-4 h-4 text-gray-700 dark:text-gray-300 hover:text-blue-800 dark:hover:text-blue-500' />
        </button>
        <span className='w-full block text-gray-500 dark:text-gray-400 text-xs border-b border-gray-300 dark:border-slate-700 font-bold px-1 pb-2'>
          <span className='truncate'>{deploymentName}</span>
        </span>
        <ul className='w-full flex flex-col gap-1 mt-2'>
          {verbs
            .sort((a, b) => Intl.Collator('en').compare(a, b))
            .map((verb) => (
              <Listbox.Option key={verb} value={verb}>
                {({ selected }) => (
                  <div
                    className={classNames(
                      'cursor-pointer inline-block',
                      !selected && 'w-full',
                      !selected && 'hover:bg-gray-200 dark:hover:bg-gray-800',
                    )}
                  >
                    <span
                      className={classNames(
                        'p-1 rounded text-sm',
                        !selected && 'text-slate-900 dark:text-slate-400',
                        selected
                          ? 'bg-blue-600 hover:bg-blue-800 text-white dark:bg-blue-500 dark:hover:bg-blue-700'
                          : '',
                      )}
                    >
                      {getNames(verb)[1]}
                    </span>
                  </div>
                )}
              </Listbox.Option>
            ))}
        </ul>
      </div>
    </li>
  )
}

export const ModulesSidebar: React.FC<{
  className?: string
  modules: Module[]
  setSelectedVerbs: React.Dispatch<React.SetStateAction<VerbId[]>>
  selectedVerbs: VerbId[]
  zoomCallbacks?: ZoomCallbacks
}> = ({ className, modules, setSelectedVerbs, selectedVerbs, zoomCallbacks }) => {
  const [query, setQuery] = useState('')

  const map: Map<string, DeploymentVerbs> = new Map()
  for (const { name: moduleName, verbs, deploymentName } of modules) {
    const value: DeploymentVerbs = {
      id: moduleName,
      deploymentName,
      verbs: new Set(),
      queriedVerbs: new Set(),
    }
    for (const { verb } of verbs) {
      verb && value.verbs.add(`${moduleName}.${verb.name}`)
    }
    map.set(moduleName, value)
  }
  const options = [...map.values()]
  const filteredOptions =
    query === ''
      ? options
      : options.reduce<DeploymentVerbs[]>((acc, option) => {
          option.queriedVerbs.clear()
          let found = option.id.toLowerCase().includes(query.toLowerCase())
          const queriedVerbs: Set<VerbId> = new Set(
            [...option.verbs].filter((verb) => verb.toLowerCase().includes(query.toLowerCase())),
          )
          if (!found) {
            found = Boolean(queriedVerbs.size)
          }
          if (!found) return acc
          option.queriedVerbs = queriedVerbs
          acc.push(option)
          return acc
        }, [])
  const handleChange: React.ChangeEventHandler<HTMLInputElement> = (event) => {
    setSelectedVerbs([])
    setQuery(event.target.value)
  }
  return (
    <div className={classNames(className, 'flex flex-col gap-2 ')}>
      <div className='px-2 shadow-sm dark:shadow-sm'>
        <input
          onChange={handleChange}
          value={query}
          className='text-sm leading-5 text-gray-900 focus:ring-0 focus:outline-none p-2 rounded w-full'
          placeholder='Search...'
        />
      </div>
      <Listbox value={selectedVerbs} onChange={setSelectedVerbs} multiple>
        <Listbox.Options static className='h-full overflow-auto flex flex-col gap-1.5 px-2'>
          {filteredOptions
            .sort((a, b) => Intl.Collator('en').compare(a.id, b.id))
            .map(({ deploymentName, id, queriedVerbs, verbs }) => {
              const displayedVerbs = query === '' ? verbs : queriedVerbs
              return (
                <ModulesOption
                  key={id}
                  verbs={[...displayedVerbs]}
                  id={id}
                  deploymentName={deploymentName}
                  zoomCallbacks={zoomCallbacks}
                />
              )
            })}
        </Listbox.Options>
      </Listbox>
    </div>
  )
}
