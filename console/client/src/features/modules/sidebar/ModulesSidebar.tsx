import React from 'react'
import { Combobox } from '@headlessui/react'
import { GetModulesResponse } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { EyeIcon } from '@heroicons/react/20/solid'
import { FilteredList } from './FilteredList'
import { ContextualList } from './ContextualList'
import { modulesContext } from '../../../providers/modules-provider'
import { VerbRef } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { Select } from '../components'

interface ModuleValue {
  deploymentName: string;
  id: string;
  type: 'module';
  name: string;
  calls?:never;
}

interface VerbValue {
  deploymentName: string;
  id: `${string}.${string}`;
  type: 'verb';
  name: string;
  calls: { key: string; value: string, type: 'verb' }[];
}

type MapValue = ModuleValue | VerbValue

export const ModulesSidebar: React.FC<{
  className: string
  setZoomID:  React.Dispatch<React.SetStateAction<`#${string}` | undefined>>
  setSelectedVerb:  React.Dispatch<React.SetStateAction< `${string}.${string}` | undefined>>
  setSelectedModule:  React.Dispatch<React.SetStateAction<string | undefined>>
  setSelectedEdges:  React.Dispatch<React.SetStateAction<`#${string}`[] | undefined>>
}> = ({
  className,
  setSelectedModule,
  setSelectedVerb,
  setZoomID,
  setSelectedEdges
}) => {
  const { modules } = React.useContext(modulesContext)
  const map: Map<string, MapValue> = new Map() 
  for(const { name: moduleName, verbs, deploymentName } of modules) {
    map.set(moduleName, { id: moduleName, name: moduleName, type: 'module', deploymentName})
    for(const {verb} of verbs) {
      if(verb) {
        const { name: verbName, metadata } = verb
        const value: VerbValue = {
          id: `${moduleName}.${verbName}`,
          type: 'verb',
          deploymentName,
          name: verbName,
          calls:[]
        }
        metadata.forEach((metadataEntry) => {
          if (metadataEntry?.value?.case === 'calls') {
            const calls = metadataEntry.value.value.calls
            calls.forEach(({ name, module}) => {
              const id = `${module}.${name}`
              value.calls.push({ key: id, value: id, type: 'verb'})
            })
          }
        })
        map.set(verbName, value)
      }       
    }
  }
  const [selected, setSelected] = React.useState<MapValue>()
  const [query, setQuery] = React.useState('')
  const options = [...map.values()]
  const handleChange = (val: MapValue) => {
    const {id, type } = val
    setSelected(val)
    if(type === 'module') {
      setSelectedModule(id)
      setSelectedVerb(undefined)
    } 
    if(type === 'verb') {
      setSelectedVerb(id)
      setSelectedModule(undefined)
    }
    setZoomID(`#${id}`)
  }

  const handleZoom: React.MouseEventHandler<HTMLButtonElement> = evt => {
    evt.stopPropagation()
    setZoomID(`#${evt.currentTarget.value}`)
  }
  const handleCall: React.MouseEventHandler<HTMLButtonElement> = evt => {
    evt.stopPropagation()
    setSe(`#${evt.currentTarget.value}`)
  }
  const filteredOptions =
    query === ''
      ? options.filter(option => option.type === 'module')
      : options.filter((option) => {
          return option.name.toLowerCase().includes(query.toLowerCase())
        })
  return (
    <div className={className}>
      <Combobox defaultValue={selected} onChange={handleChange} by="id">
         <Combobox.Input
          onChange={(event) => setQuery(event.target.value)}
          displayValue={(option:MapValue ) => option.name}
          className='text-sm leading-5 text-gray-900 focus:ring-0'
        />
        <Combobox.Options static>
          {filteredOptions.map((option) => (
            <Combobox.Option
              key={option.id}
              value={option}
            >
              <span>{option.type}</span> {option.name}
              {option.type === 'module'  && <button
                value={option.id}
                onClick={handleZoom}
                >
                  <EyeIcon className='text-gray-70 w-3 height-3'/>
              </button>}
              <span>{option.deploymentName}</span>
            </Combobox.Option>
          ))}
        </Combobox.Options>
      </Combobox>
    </div>
    
  )
}