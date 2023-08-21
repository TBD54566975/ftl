import { ChevronRightIcon } from '@heroicons/react/20/solid'
import { useContext } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { modulesContext } from '../../providers/modules-provider'
import { getCodeBlock } from '../../utils/data.utils'
import { classNames } from '../../utils/react.utils'
import { getCalls, getVerbCode } from './verb.utils'
import { VerbCalls } from './VerbCalls.tsx'
import { VerbForm } from './VerbForm.tsx'
import { syntaxTheme } from '../../utils/style.utils.ts'

export default function VerbPage() {
  const { moduleId, id } = useParams()
  const modules = useContext(modulesContext)

  const module = modules.modules.find(m => m.name === moduleId)
  const verb = module?.verbs.find(v => v.verb?.name === id?.toLocaleLowerCase())
  const callData = module?.data.filter(data =>
    [ verb?.verb?.request?.name, verb?.verb?.response?.name ].includes(data.name)
  )

  if (module === undefined || verb === undefined) {
    return <></>
  }

  return (
    <div className='min-w-0 flex-auto p-4'>
      <nav className='flex'
        aria-label='Breadcrumb'
      >
        <ol role='list'
          className='flex items-center space-x-4'
        >
          <li key='/modules'>
            <div>
              <Link to='/modules'
                className='text-sm font-medium text-gray-400 hover:text-gray-500'
              >
                Modules
              </Link>
            </div>
          </li>
          <li key={module.name}>
            <div className='flex items-center'>
              <ChevronRightIcon className='h-5 w-5 flex-shrink-0 text-gray-400'
                aria-hidden='true'
              />
              <Link
                to={`/modules/${module.name}`}
                className='ml-4 text-sm font-medium text-gray-400 hover:text-gray-500'
                aria-current={'page'}
              >
                {module.name}
              </Link>
            </div>
          </li>
        </ol>
      </nav>
      <div className='text-sm pt-4'>
        <SyntaxHighlighter language='go'
          style={syntaxTheme}
        >
          {getVerbCode(verb?.verb)}
        </SyntaxHighlighter>
      </div>
      <div className='pt-4'>
        {callData?.map((data, index) => (
          <div key={index}
            className='text-sm'
          >
            <SyntaxHighlighter language='go'
              style={syntaxTheme}
            >
              {getCodeBlock(data)}
            </SyntaxHighlighter>
          </div>
        ))}
      </div>
      <VerbForm
        module={module}
        verb={verb}
      />
      <div className='flex items-center gap-x-3 pt-6'>
        <h2 className='min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white'>
          <div className='flex gap-x-2'>
            <span className='truncate'>Calls</span>
          </div>
        </h2>
      </div>
      {getCalls(verb?.verb).map(call =>
        call.calls.map(call => (
          <Link key={`/modules/${call.module}/verbs/${call.name}`}
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

      <VerbCalls module={module} verb={verb} />

      <div className='flex items-center gap-x-3 pt-6'>
        <h2 className='min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white'>
          <div className='flex gap-x-2'>
            <span className='truncate'>Errors</span>
          </div>
        </h2>
      </div>
    </div>
  )
}
