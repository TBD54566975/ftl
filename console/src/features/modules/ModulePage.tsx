import { useParams } from 'react-router-dom'
import { modules } from '../../data/Modules'
import { classNames } from '../../utils'
import { environments, statuses } from '../../data/Types'
import ModuleNotFound from './ModuleNotFound'
import VerbList from '../verbs/VerbList'

export default function ModulePage() {
  const { id } = useParams()
  const module = modules.find(module => module.id === id?.toLocaleLowerCase())

  if (module === undefined) {
    return <ModuleNotFound id={id} />
  }

  return (
    <>
      <div className="relative flex items-center space-x-4">
        <div className="min-w-0 flex-auto">
          <div className="flex items-center gap-x-3">
            <div
              className={classNames(
                statuses[module.status],
                'flex-none rounded-full p-1'
              )}
            >
              <div className="h-2 w-2 rounded-full bg-current" />
            </div>
            <h2 className="min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white">
              <div className="flex gap-x-2">
                <span className="truncate">{module.name}</span>
                <span className="text-gray-400">/</span>
                <span className="whitespace-nowrap">{module.language}</span>
                <span className="absolute inset-0" />
              </div>
            </h2>
          </div>
          <div className="mt-3 flex items-center gap-x-2.5 text-xs leading-5 text-gray-400">
            <p className="truncate">{module.description}</p>
            <svg
              viewBox="0 0 2 2"
              className="h-0.5 w-0.5 flex-none fill-gray-300"
            >
              <circle cx={1} cy={1} r={1} />
            </svg>
            <p className="whitespace-nowrap">{module.statusText}</p>
          </div>
        </div>
        <div
          className={classNames(
            environments[module.environment],
            'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
          )}
        >
          {module.environment}
        </div>
      </div>
      <h2 className="text-base font-semibold dark:text-white py-4">Verbs</h2>
      <VerbList module={module} />
    </>
  )
}
