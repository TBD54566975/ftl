import { useContext } from 'react'
import { Link } from 'react-router-dom'
import { Card } from '../../components/Card'
import { modulesContext } from '../../providers/modules-provider'
import { classNames } from '../../utils/react.utils'
import { statuses } from '../../utils/style.utils'

export default function ModulesPage() {
  const modules = useContext(modulesContext)

  return (
    <>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {modules.modules.map(module => (
          <Card key={module.name}>
            <div className="min-w-0 flex-1">
              <Link to={`${module.name}`} className="focus:outline-none">
                <span className="absolute inset-0" aria-hidden="true" />
                <div className="min-w-0 flex-auto">
                  <div className="flex items-center gap-x-3">
                    <div className={classNames(statuses['online'], 'flex-none rounded-full p-1')}>
                      <div className="h-2 w-2 rounded-full bg-current" />
                    </div>
                    <p className="text-sm font-medium text-gray-900 dark:text-gray-300">{module.name}</p>
                    <div className="rounded-md bg-gray-100 dark:bg-gray-700/40 px-2 py-1 text-xs font-medium text-gray-600 dark:text-gray-400 ring-1 ring-inset ring-black/10 dark:ring-white/10">
                      {module.language}
                    </div>
                  </div>
                </div>
              </Link>
            </div>
          </Card>
        ))}
        {/* {schema.map(module => (
          <Card key={module.schema?.name}>
            <div className="min-w-0 flex-1">
              <Link to={`${module.schema?.name}`} className="focus:outline-none">
                <span className="absolute inset-0" aria-hidden="true" />
                <div className="min-w-0 flex-auto">
                  <div className="flex items-center gap-x-3">
                    <div className={classNames(statuses['online'], 'flex-none rounded-full p-1')}>
                      <div className="h-2 w-2 rounded-full bg-current" />
                    </div>
                    <p className="text-sm font-medium text-gray-900 dark:text-gray-300">{module.schema?.name}</p>
                  </div>
                </div>

                {(module.schema?.comments.length ?? 0) > 0 && (
                  <div className="min-w-0 flex-auto pt-2">
                    <div className="flex items-center gap-x-3">
                      <p className="truncate text-sm text-gray-500">{module.schema?.comments}</p>
                    </div>
                  </div>
                )}
              </Link>
            </div>
          </Card>
        ))} */}
      </div>
    </>
  )
}
