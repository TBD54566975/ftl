import { useContext } from 'react'
import { SelectedModuleContext } from '../../providers/selected-module-provider'
import { Link } from 'react-router-dom'
import { textColor } from '../../utils/style.utils'

export function ModuleDetails() {
  const { selectedModule } = useContext(SelectedModuleContext)

  if (!selectedModule) {
    return (
      <div className='flex-1 p-4 overflow-auto flex items-center justify-center'>
        <span>No module selected</span>
      </div>
    )
  }

  return (
    <div className='flex-1 overflow-auto text-sm font-medium text-gray-400'>
      <div className='flex justify-between'>
        <dt>Name</dt>
        <dd className={`${textColor}`}>{selectedModule.name}</dd>
      </div>
      <div className='flex pt-2 justify-between'>
        <dt>Deployment</dt>
        <dd className={`${textColor}`}>{selectedModule.deploymentName}</dd>
      </div>
      <div className='flex pt-2 justify-between'>
        <dt>Language</dt>
        <dd className={`${textColor}`}>{selectedModule.language}</dd>
      </div>

      <div className='flex pt-4 justify-between'>
        <dt>Verbs</dt>
        <dd className='text-white flex flex-col space-y-2'>
          {selectedModule.verbs.map((verb, index) => (
            <Link
              key={index}
              to={`/modules/${selectedModule.name}/verbs/${verb.verb?.name}`}
              className='rounded bg-indigo-600 px-2 py-1 text-xs font-semibold text-white text-center
            shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'
            >
              {verb.verb?.name}
            </Link>
          ))}
        </dd>
      </div>

      <div className='flex pt-4 justify-between text-right'>
        <dt>Data</dt>
        <dd className='text-white'>
          <ul className='list-none ml-4'>
            {selectedModule.data.map((data, index) => (
              <li key={index} className={`${textColor}`}>
                <code>{data.name}</code>
              </li>
            ))}
          </ul>
        </dd>
      </div>
    </div>
  )
}
