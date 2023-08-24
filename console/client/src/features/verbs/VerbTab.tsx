import React from 'react'
import { CodeBlock } from '../../components/CodeBlock'
import { VerbCalls } from './VerbCalls'
import { VerbForm } from './VerbForm'
import { getVerbCode } from './verb.utils'
import { modulesContext } from '../../providers/modules-provider'
type Props = {
  id: string
}

export const VerbTab: React.FC<Props> = ({ id }) => {
  const [ moduleId, verbName ] = id.split('.')
  const modules = React.useContext(modulesContext)
  const module = modules.modules.find(module => module?.name === moduleId)
  const verb = module?.verbs.find(v => v.verb?.name === verbName)
  const callData = module?.data.filter(data =>
    [ verb?.verb?.request?.name, verb?.verb?.response?.name ].includes(data.name)
  )

  if (!module || !verb) {
    return <></>
  }

  return (
    <div className='min-w-0 flex-auto p-4'>
      <VerbForm module={module} verb={verb} />
      <CodeBlock code={getVerbCode(verb?.verb)} language='go' />

      <div className='pt-4'>
        {callData?.map((data, index) => (
          <CodeBlock key={index} code={getVerbCode(data)} language='go' />
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
