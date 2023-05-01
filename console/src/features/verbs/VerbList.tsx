import { ChevronRightIcon } from '@heroicons/react/20/solid'
import { classNames } from '../../utils'
import { environments } from '../../data/Types'

export default function VerbList({ module }) {
  return (
    <>
      <ul role="list" className="divide-y divide-black/5 dark:divide-white/5">
        {module.verbs.map(verb => (
          <li
            key={verb.id}
            className="relative flex items-center space-x-4 py-4"
          >
            <div className="min-w-0 flex-auto">
              <div className="flex items-center gap-x-3">
                <h2 className="min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white">
                  <a
                    href={`${module.id}/verbs/${verb.id}`}
                    className="flex gap-x-2"
                  >
                    <span className="truncate">{verb.name}</span>
                    <span className="absolute inset-0" />
                  </a>
                </h2>
              </div>
              <div className="mt-3 flex items-center gap-x-2.5 text-xs leading-5 text-gray-400">
                <p className="truncate">{verb.description}</p>
              </div>
            </div>
            <div
              className={classNames(
                environments[module.environment],
                'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
              )}
            >
              {verb.id}
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
