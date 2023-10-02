import React from 'react'
import { Combobox, Listbox, Disclosure} from '@headlessui/react'
import { EyeIcon, ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/20/solid'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbId } from './modules.constants'
import { getVerbName } from './modules.utils'

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
  deploymentName,
  selectedVerbs
}) => {
  // const hasVerbs = verbs.size
  // const [isChecked, setIsChecked] = React.useState(false);
  // const [checkedVerbs, setCheckedVerbs] = React.useState<Set<VerbId>>(new Set())
  // const handleChange: React.ChangeEventHandler<HTMLInputElement> = (event) => {
  //   event.preventDefault()
  //   const selected = event.target.checked
  //   setIsChecked(selected);
  //   setSelectedModules(modules => {
  //     modules.has(id) && !selected
  //       ? modules.delete(id)
  //       : modules.add(id)
  //     return modules
  //   })
  // }
  // const handleSelectAll: React.MouseEventHandler<HTMLButtonElement> = (event) => {
  //   event.preventDefault()
  //   setSelectedVerbs(selectedVerbs => {
  //     verbs.forEach(verb => selectedVerbs.add(verb))
  //     return selectedVerbs
  //   })
  //   setCheckedVerbs(verbs)
  // }
  // const handleSelectNone: React.MouseEventHandler<HTMLButtonElement> = (event) => {
  //   event.preventDefault()
  //   setSelectedVerbs(selectedVerbs => {
  //     verbs.forEach(verb => selectedVerbs.delete(verb))
  //     return selectedVerbs
  //   })
  //   setCheckedVerbs(checkedVerbs => {
  //     checkedVerbs.clear()
  //     return checkedVerbs
  //   })
  // }
  // const handleVerbChange: React.ChangeEventHandler<HTMLInputElement> = (event) => {
  //   event.preventDefault()
  //   const selected = event.target.checked
  //   const value =  event.target.value as VerbId
  //   if(selected) {
  //     setSelectedVerbs(selectedVerbs => {
  //       selectedVerbs.add(value)
  //       return selectedVerbs
  //     })
  //     setCheckedVerbs(checkedVerbs => {
  //       checkedVerbs.add(value)
  //       return checkedVerbs
  //     })
  //   } else {
  //     setSelectedVerbs(selectedVerbs => {
  //       selectedVerbs.delete(value)
  //       return selectedVerbs
  //     })
  //     setCheckedVerbs(checkedVerbs => {
  //       checkedVerbs.delete(value)
  //       return checkedVerbs
  //     })
  //   }
  // }
  return (
    <Listbox.Option value={id}>
      <span>{id}</span>
      <button onClick={() => setZoomId(id)}><EyeIcon className='w-3 h-3'/></button>
      <Listbox value={selectedVerbs} onChange={setSelectedVerbs} multiple>
        <Listbox.Options static>
          {verbs.map((verb) => (
            <Listbox.Option key={verb} value={verb}>
              {getVerbName(verb)}
            </Listbox.Option>
          ))}
        </Listbox.Options>
      </Listbox>
    </Listbox.Option>
   
  )
  // return (
  //   <div>
  //     <label>
  //       {id}
  //       <input type="checkbox" onChange={handleChange} value={id} checked={isChecked}/>
  //     </label>
  //     <button type="button" onClick={handleZoom} value={id}><EyeIcon className='w-3 h-3'/></button>
  //     <span>{deploymentName}</span>
  //     {hasVerbs && (<>
  //       <button type="button" onClick={handleSelectAll}>
  //         All
  //       </button>
  //       <span> | </span>
  //       <button type="button" onClick={handleSelectNone}>
  //         None
  //       </button>
  //     </>)}
  //     {isChecked && (
  //       <ul>
  //         {[...verbs].map(verb => (
  //           <li key={verb}>
  //             <label>
  //               <input type="checkbox" value={verb} onChange={handleVerbChange} checked={checkedVerbs.has(verb)}/>
  //               <span>{getVerbName(verb)}</span>
  //             </label>
  //           </li>
  //         ))}
  //       </ul>
  //     )}
  //   </div>
  // )
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
    <div className={className}>
      <input
        onChange={handleChange}
        value={query}
        className='text-sm leading-5 text-gray-900 focus:ring-0'
      />
       <Listbox value={selectedModules} onChange={setSelectedModules} multiple>
        <Listbox.Options static>
          {filteredOptions.map(({ deploymentName, id, queriedVerbs, verbs}) => {
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