import React from 'react'
import { Combobox, Listbox, Disclosure} from '@headlessui/react'
import { EyeIcon, ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/20/solid'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbId } from './modules.constants'
import { getVerbName } from './modules.utils'
import { classNames } from '../../utils'

interface MapValue {
  deploymentName: string;
  id: string;
  verbs: Set<VerbId>
  queriedVerbs: Set<VerbId>
}

const ModulesOption: React.FC<{
  id: string
  verbs: VerbId[]
  setSelectedVerbs: React.Dispatch<React.SetStateAction<VerbId[]>>
  selectedVerbs: VerbId[]
  setZoomId:  React.Dispatch<React.SetStateAction<string | undefined>>
  deploymentName: string
}> = ({
  id,
  setSelectedVerbs,
  setZoomId,
  verbs,
  selectedVerbs
}) => {
  return (
    <Listbox.Option value={id}>
      <span>{id}</span>
      <button onClick={() => setZoomId(id)}><EyeIcon className='w-3 h-3'/></button>
      <Listbox value={selectedVerbs} onChange={setSelectedVerbs} multiple>
        <Listbox.Options static>
          {verbs
            .sort((a, b) => Intl.Collator('en').compare(a, b))
            .map((verb) => (
              <Listbox.Option key={verb} value={verb}>
                {getVerbName(verb)}
              </Listbox.Option>
            ))
          }
        </Listbox.Options>
      </Listbox>
    </Listbox.Option>
   
  )
}

export const ModulesSidebar: React.FC<{
  className: string
  modules: Module[]
  setSelectedVerbs:  React.Dispatch<React.SetStateAction<VerbId[]>>
  setSelectedModules:  React.Dispatch<React.SetStateAction<string[]>>
  selectedModules: string[],
  selectedVerbs: VerbId[]
  setZoomId: React.Dispatch<React.SetStateAction<string | undefined>>
}> = ({
  className,
  modules,
  setSelectedVerbs,
  setSelectedModules,
  selectedModules,
  selectedVerbs,
  setZoomId
}) => {
  /** Setup hooks */
  const [query, setQuery] = React.useState('')

  /** Format data */
  const map: Map<string, MapValue> = new Map() 
  for(const { name: moduleName, verbs, deploymentName } of modules) {
    const value: MapValue = {
      id: moduleName,
      deploymentName,
      verbs: new Set(),
      queriedVerbs: new Set()
    }
    for(const {verb} of verbs) {
      verb && value.verbs.add(`${moduleName}.${verb.name}`)
    }
    map.set(moduleName, value)
  }
  const options = [...map.values()]
  const filteredOptions = query === ''
      ? options
      : options.reduce<MapValue[]>((acc, option) => {
          option.queriedVerbs.clear()
          let found =  option.id.toLowerCase().includes(query.toLowerCase())
          const queriedVerbs: Set<VerbId> =new Set([...option.verbs].filter(verb => verb.toLowerCase().includes(query.toLowerCase())))
          if(!found) {
            found = Boolean(queriedVerbs.size)
          }
          if(!found) return acc
          option.queriedVerbs = queriedVerbs
          acc.push(option)
          return acc
        }, [])
  const handleChange: React.ChangeEventHandler<HTMLInputElement> = event => {
    setSelectedModules([])
    setSelectedVerbs([])
    setQuery(event.target.value)
  }
  return (
    <div className={classNames(className, 'flex flex-col')}>
      <input
        onChange={handleChange}
        value={query}
        className='text-sm leading-5 text-gray-900 focus:ring-0'
      />
       <Listbox value={selectedModules} onChange={setSelectedModules} multiple>
        <Listbox.Options static className='flex-1 overflow-auto'>
          {filteredOptions
            .sort((a, b) => Intl.Collator('en').compare(a.id, b.id))
            .map(({ deploymentName, id, queriedVerbs, verbs}) => {
            const displayedVerbs = query === ''
              ? verbs
              : queriedVerbs
            return (
              <ModulesOption 
                key={id}
                verbs={[...displayedVerbs]}
                id={id}
                deploymentName={deploymentName}
                setSelectedVerbs={setSelectedVerbs}
                selectedVerbs={selectedVerbs}
                setZoomId={setZoomId}
              />
            )
          })}
        </Listbox.Options>
       </Listbox>
    </div>
  )
}