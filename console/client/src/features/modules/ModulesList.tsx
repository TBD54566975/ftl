import { useContext } from 'react'
import { modulesContext } from '../../providers/modules-provider'

export function ModulesList() {
  const modules = useContext(modulesContext)
  return (
    <ul role='list'
      className='space-y-2'
    >
      {modules.modules?.map(module => (
        <li key={module.deploymentName}
          className='relative flex gap-x-4 p-2 rounded cursor-pointer shadow-sm border border-transparent
        bg-slate-200 hover:bg-indigo-700 hover:text-white hover:border-indigo-600
        dark:bg-slate-800 dark:hover:bg-indigo-700 dark:hover:text-white dark:hover:border-indigo-600'
        >
          <div className='flex-1 truncate'>
            {module.name}
          </div>
          <div>
            <div className={`inline-block rounded-md dark:bg-gray-700/40 px-2 py-1 text-xs font-medium text-gray-500 dark:text-gray-400 ring-1 ring-inset ring-black/10 dark:ring-white/10`}>
              {module.deploymentName}
            </div>
          </div>
        </li>
      ))}
    </ul>
  )
}
