import { Link } from 'react-router-dom'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { getCodeBlock } from '../../utils/data.utils'
import { classNames } from '../../utils/react.utils'
import { getCalls } from './verb.utils'
import { VerbCalls } from './VerbCalls.tsx'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb.ts'

export const VerbPage: React.FC<{module: Module, id}> = ({ module, id }) =>{
  const verb = module?.verbs.find(v => v.verb?.name === id?.toLocaleLowerCase())
  const callData = module?.data.filter(data =>
    [ verb?.verb?.request?.name, verb?.verb?.response?.name ].includes(data.name)
  )

  if (module === undefined || verb === undefined) {
    return <></>
  }

  return (
    <div className='min-w-0 flex-auto'>
      <div className='pt-4'>
        {callData?.map(data => (
          <div key={data.name}
            className='text-sm'
          >
            <SyntaxHighlighter language='go'
              style={atomDark}
            >
              {getCodeBlock(data)}
            </SyntaxHighlighter>
          </div>
        ))}
      </div>
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

      <VerbCalls module={module}
        verb={verb}
      />

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
