import React from 'react'
import { schemaContext } from '../../providers/schema-provider.tsx'
import { classNames } from '../../utils/react.utils'
import { statuses } from '../../utils/style.utils'
import { useNavigate , useLocation } from 'react-router-dom'
import { modulesContext } from '../../providers/modules-provider.tsx'
import { ModuleTimeline } from './ModuleTimeline.tsx'
import * as styles from './Module.module.css'

export default function ModulesPage() {
  const schema = React.useContext(schemaContext)
  const modules = React.useContext(modulesContext)
  const  [ id, setId ] = React.useState<string | undefined>(undefined)
  const module = modules.modules.find(module => module?.name === id)

  const navigate = useNavigate()
  const location = useLocation()
  return (
    <div className={styles.grid}>
      <div className={ styles.filter}>
        {schema.map(module => {
          const name = module.schema?.name
          return (
            <button
              className={`relative flex items-center space-x-3 rounded-lg border border-gray-300 bg-white dark:bg-slate-800 dark:border-indigo-400 px-6 py-5 shadow-sm focus-within:ring-2 focus-within:ring-indigo-500 focus-within:ring-offset-2 dark:focus-within:ring-2 dark:focus-within:ring-indigo-400 dark:focus-within:ring-offset-2 hover:border-gray-400 dark:hover:border-indigo-200`}
              key={name}
              onClick={() => {
                if(!name) return
                const searchParams = new URLSearchParams(location.search)
                searchParams.set('module', name)
                navigate({ ...location, search: searchParams.toString() })
                setId(name)
              }}
            >
              <div className='min-w-0 flex-1'>
                <span className='absolute inset-0'
                  aria-hidden='true'
                />
                <div className='min-w-0 flex-auto'>
                  <div className='flex items-center gap-x-3'>
                    <div className={classNames(statuses['online'], 'flex-none rounded-full p-1')}>
                      <div className='h-2 w-2 rounded-full bg-current' />
                    </div>
                    <p className='text-sm font-medium text-gray-900 dark:text-gray-300'>{name}</p>
                  </div>
                  <div className='pt-4'>
                    <div className={`inline-block rounded-md dark:bg-gray-700/40 px-2 py-1 text-xs font-medium text-gray-500 dark:text-gray-400 ring-1 ring-inset ring-black/10 dark:ring-white/10`}>
                      {module.deploymentKey}
                    </div>
                  </div>
                </div>

                {(module.schema?.comments.length ?? 0) > 0 && (
                  <div className='min-w-0 flex-auto pt-2'>
                    <div className='flex items-center gap-x-3'>
                      <p className='truncate text-sm text-gray-500'>{module.schema?.comments}</p>
                    </div>
                  </div>
                )}
              </div>
            </button>
          )})}
      </div>
      <ModuleTimeline module={module} />
      {/* <div className={styles.misc}>
        <VerbList module={module} />
      </div> */}
    </div>
  )
}
