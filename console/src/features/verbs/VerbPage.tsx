import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { Link, useParams } from 'react-router-dom'
import { ChevronRightIcon } from '@heroicons/react/20/solid'
import { Verb } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { useContext } from 'react'
import { schemaContext } from '../../providers/schema-provider'
import { classNames } from '../../utils/react.utils'
import { getData } from '../modules/module.utils'
import { getCodeBlock } from '../../utils/data.utils'
import { getCalls, getVerbCode } from './verb.utils'

export default function VerbPage() {
  const { moduleId, id } = useParams()
  const schema = useContext(schemaContext)
  const module = schema.find(module => module.schema?.name === moduleId)?.schema

  const verb = module?.decls.find(
    decl => decl.value.case === 'verb' && decl.value.value.name === id?.toLocaleLowerCase(),
  )?.value.value as Verb

  const callData = getData(module).filter(data => [verb.request?.name, verb.response?.name].includes(data.name))

  if (module === undefined || verb === undefined) {
    return <></>
  }

  return (
    <div className="min-w-0 flex-auto">
      <nav className="flex" aria-label="Breadcrumb">
        <ol role="list" className="flex items-center space-x-4">
          <li key="/modules">
            <div>
              <Link to="/modules" className="text-sm font-medium text-gray-400 hover:text-gray-500">
                Modules
              </Link>
            </div>
          </li>
          <li key={module.name}>
            <div className="flex items-center">
              <ChevronRightIcon className="h-5 w-5 flex-shrink-0 text-gray-400" aria-hidden="true" />
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
      <div className="text-sm pt-4">
        <SyntaxHighlighter language="go" style={atomDark}>
          {getVerbCode(verb)}
        </SyntaxHighlighter>
      </div>

      <div className="pt-4">
        {callData.map(data => (
          <div key={data.name} className="text-sm">
            <SyntaxHighlighter language="go" style={atomDark}>
              {getCodeBlock(data)}
            </SyntaxHighlighter>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-x-3 pt-6">
        <h2 className="min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white">
          <div className="flex gap-x-2">
            <span className="truncate">Calls</span>
          </div>
        </h2>
      </div>

      {getCalls(verb).map(call =>
        call.calls.map(call => (
          <Link key={`/modules/${call.module}/verbs/${call.name}`} to={`/modules/${call.module}/verbs/${call.name}`}>
            <span
              className={classNames(
                'text-indigo-400 bg-indigo-400/10 ring-indigo-400/30',
                'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset',
              )}
            >
              {call.name}
            </span>
          </Link>
        )),
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
