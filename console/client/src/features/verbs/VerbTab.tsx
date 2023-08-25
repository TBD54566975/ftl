import React from 'react'
import {CodeBlock} from '../../components/CodeBlock'
import {modulesContext} from '../../providers/modules-provider'
import {VerbCalls} from './VerbCalls'
import {VerbForm} from './VerbForm'
import {buildVerbSchema} from './verb.utils'
type Props = {
  id: string
}

export const VerbTab: React.FC<Props> = ({id}) => {
  const [moduleId, verbName] = id.split('.')
  const modules = React.useContext(modulesContext)
  const module = modules.modules.find(module => module?.name === moduleId)
  const verb = module?.verbs.find(v => v.verb?.name === verbName)

  const callData =
    module?.data.filter(data =>
      [verb?.verb?.request?.name, verb?.verb?.response?.name].includes(
        data.data?.name
      )
    ) ?? []

  if (!module || !verb?.verb) {
    return <></>
  }

  return (
    <div className='min-w-0 flex-auto p-4'>
      <VerbForm
        module={module}
        verb={verb}
      />

      <div className='pt-4'>
        <CodeBlock
          code={buildVerbSchema(
            verb?.schema,
            callData.map(d => d.schema)
          )}
          language='graphql'
        />
      </div>

      <VerbCalls
        module={module}
        verb={verb}
      />
    </div>
  )
}
