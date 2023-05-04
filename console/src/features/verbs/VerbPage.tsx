import { Link, useParams } from 'react-router-dom'
import { ChevronRightIcon } from '@heroicons/react/20/solid'
import {
  MetadataCalls,
  Verb
} from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { classNames } from '../../utils'
import { useContext } from 'react'
import { schemaContext } from '../../providers/schema-provider'

export default function VerbPage() {
  const { moduleId, id } = useParams()
  const schema = useContext(schemaContext)
  const module = schema.find(module => module.schema?.name === moduleId)?.schema

  const verb = module?.decls.find(
    decl =>
      decl.value.case === 'verb' &&
      decl.value.value.name === id?.toLocaleLowerCase()
  )?.value.value as Verb

  const calls = verb?.metadata
    .filter(meta => meta.value.case === 'calls')
    .map(meta => meta.value.value as MetadataCalls)

  if (module === undefined || verb === undefined) {
    return <></>
  }

  return (
    <div className="min-w-0 flex-auto">
      <nav className="flex" aria-label="Breadcrumb">
        <ol role="list" className="flex items-center space-x-4">
          <li>
            <div>
              <Link
                to="/modules"
                className="text-sm font-medium text-gray-400 hover:text-gray-500"
              >
                Modules
              </Link>
            </div>
          </li>
          <li key={module.name}>
            <div className="flex items-center">
              <ChevronRightIcon
                className="h-5 w-5 flex-shrink-0 text-gray-400"
                aria-hidden="true"
              />
              <Link
                to={`/modules/${module.name}`}
                className="ml-4 text-sm font-medium text-gray-400 hover:text-gray-500"
                aria-current={'page'}
              >
                {module.name}
              </Link>
            </div>
          </li>
        </ol>
      </nav>
      <div className="flex items-center gap-x-3 pt-6">
        <h2 className="min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white">
          <div className="flex gap-x-2">
            <span className="truncate">{verb.name}</span>
          </div>
        </h2>
      </div>
      <div className="relative flex items-center space-x-4">
        <div className="min-w-0 flex-auto text-indigo-500 dark:text-indigo-400">
          <code className="text-sm">
            {`${verb.name}(${verb.request?.name}) -> ${verb.response?.name}`}
          </code>
        </div>
      </div>

      <div className="flex items-center gap-x-3 pt-6">
        <h2 className="min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white">
          <div className="flex gap-x-2">
            <span className="truncate">Calls</span>
          </div>
        </h2>
      </div>

      {calls?.map(call =>
        call.calls.map(call => (
          <Link
            key={`/modules/${call.module}/verbs/${call.name}`}
            to={`/modules/${call.module}/verbs/${call.name}`}
          >
            <span
              className={classNames(
                'text-indigo-400 bg-indigo-400/10 ring-indigo-400/30',
                'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
              )}
            >
              {call.name}
            </span>
          </Link>
        ))
      )}
      <div className="flex items-center gap-x-3 pt-6">
        <h2 className="min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white">
          <div className="flex gap-x-2">
            <span className="truncate">Errors</span>
          </div>
        </h2>
      </div>
    </div>
  )
}
