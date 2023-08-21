import { useContext } from 'react'
import { SelectedModuleContext } from '../../providers/selected-module-provider'
import { Link } from 'react-router-dom'

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
    <div className='flex-1 p-2 overflow-auto dark:bg-gray-900'>
      <div className='mb-2'>
        <strong className='text-gray-600 dark:text-gray-400'>Name:</strong>
        <span className='ml-2 dark:text-white'>{selectedModule.name}</span>
      </div>

      <div className='mb-2'>
        <strong className='text-gray-600 dark:text-gray-400'>Deployment Name:</strong>
        <span className='ml-2 dark:text-white'>{selectedModule.deploymentName}</span>
      </div>

      <div className='mb-2'>
        <strong className='text-gray-600 dark:text-gray-400'>Language:</strong>
        <span className='ml-2 dark:text-white'>{selectedModule.language}</span>
      </div>

      <div className='mb-2'>
        <strong className='text-gray-600 dark:text-gray-400'>Verbs:</strong>
        <ul className='list-disc ml-4'>
          {selectedModule.verbs.map((verb, index) => (
            <li key={index}
              className='dark:text-white'
            >
              <Link to={`/modules/${selectedModule.name}/verbs/${verb?.verb?.name}`}>{verb.verb?.name}</Link>
            </li>
          ))}
        </ul>
      </div>

      <div className='mb-2'>
        <strong className='text-gray-600 dark:text-gray-400'>Data:</strong>
        <ul className='list-disc ml-4'>
          {selectedModule.data.map((data, index) => (
            <li key={index}
              className='dark:text-white'
            >{data.name}</li>
          ))}
        </ul>
      </div>

    </div>
  )
}
