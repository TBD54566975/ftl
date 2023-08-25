import React from 'react'
import { modulesContext } from '../../providers/modules-provider'
import { SelectedModuleContext } from '../../providers/selected-module-provider'
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom'

export function ModulesList() {
  const modules = React.useContext(modulesContext)
  const { selectedModule, setSelectedModule } = React.useContext(SelectedModuleContext)
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams] = useSearchParams()
  React.useEffect(() => {
    const id = searchParams.get('details')
    const module = modules.modules.find((module) => module?.name === id)
    module && setSelectedModule(module)
  })
  return (
    <ul role='list' className='space-y-2'>
      {modules.modules?.map((module) => (
        <li
          key={module.deploymentName}
          onClick={() => {
            setSelectedModule((prevModule) => (prevModule === module ? null : module))
            searchParams.set('details', module.name)
            navigate({ ...location, search: searchParams.toString() })
          }}
          className={`relative flex gap-x-4 p-2 rounded cursor-pointer shadow-sm border border-transparent
          ${
            module === selectedModule
              ? 'bg-indigo-700 text-white dark:bg-indigo-700 dark:text-white'
              : 'bg-slate-100 hover:bg-indigo-700 hover:text-white hover:border-indigo-600 dark:bg-slate-800 dark:hover:bg-indigo-700 dark:hover:text-white dark:hover:border-indigo-600'
          }
          dark:hover:bg-indigo-700 dark:hover:text-white dark:hover:border-indigo-600`}
        >
          <div className='flex-1 truncate'>{module.name}</div>
          <div>
            <div
              className={`inline-block rounded-md dark:bg-gray-700/40 px-2 py-1 text-xs font-medium
            ${module === selectedModule ? 'text-gray-300' : 'text-gray-500 dark:text-gray-400'}
             ring-1 ring-inset ring-black/10 dark:ring-white/10`}
            >
              {module.deploymentName}
            </div>
          </div>
        </li>
      ))}
    </ul>
  )
}
