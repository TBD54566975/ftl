import React from 'react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { syntaxTheme } from '../../utils/style.utils'
import { getVerbCode } from './verb.utils'
import { getCodeBlock } from '../../utils/data.utils'
import { VerbForm } from './VerbForm'
import { VerbCalls } from './VerbCalls'
import { modulesContext } from '../../providers/modules-provider'

type Props = {
  id: string
}

export const VerbTab: React.FC<Props> = ({ id }) => {
  const [ moduleId, verbName ] = id.split('.')
  const modules = React.useContext(modulesContext)
  const module = modules.modules.find(module => module?.name === moduleId)
  const verb = module?.verbs.find(v => v.verb?.name === verbName?.toLocaleLowerCase())
  const callData = module?.data.filter(data =>
    [ verb?.verb?.request?.name, verb?.verb?.response?.name ].includes(data.name)
  )

  if (!module || !verb) {
    return <></>
  }

  return (
    <div className='min-w-0 flex-auto p-4'>
      <VerbForm module={module} verb={verb} />
      <div className='text-sm pt-4'>
        <SyntaxHighlighter language='go'
          style={syntaxTheme()}
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
              style={syntaxTheme()}
            >
              {getCodeBlock(data)}
            </SyntaxHighlighter>
          </div>
        ))}
      </div>

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
