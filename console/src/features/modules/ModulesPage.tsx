import { ChevronRightIcon } from '@heroicons/react/20/solid'
import { classNames } from '../../utils'
import { environments, statuses } from '../../data/Types'
import { useSchema } from '../../hooks/use-schema'

export default function ModulesPage() {
  const schema = useSchema()

  return (
    <>
      <h2 className="text-base font-semibold dark:text-white">Modules</h2>
      <ul role="list" className="divide-y divide-black/5 dark:divide-white/5">
        {schema.map(module => (
          <li
            key={module.schema?.name}
            className="relative flex items-center space-x-4 py-4"
          >
            <div className="min-w-0 flex-auto">
              <div className="flex items-center gap-x-3">
                <div
                  className={classNames(
                    statuses['online'],
                    'flex-none rounded-full p-1'
                  )}
                >
                  <div className="h-2 w-2 rounded-full bg-current" />
                </div>
                <h2 className="min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white">
                  <a
                    href={`modules/${module.schema?.name}`}
                    className="flex gap-x-2"
                  >
                    <span className="truncate">{module.schema?.name}</span>
                    <span className="text-gray-400">/</span>
                    <span className="whitespace-nowrap">go</span>
                    <span className="absolute inset-0" />
                  </a>
                </h2>
              </div>
            </div>
            <div
              className={classNames(
                environments['Staging'],
                'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
              )}
            >
              Staging
            </div>
            <ChevronRightIcon
              className="h-5 w-5 flex-none text-gray-400"
              aria-hidden="true"
            />
          </li>
        ))}
      </ul>
    </>
  )
}
