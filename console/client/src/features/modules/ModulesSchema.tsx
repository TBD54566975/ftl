import React from 'react'
import { Tab } from '@headlessui/react'
import { classNames } from '../../utils'
import { buildVerbSchema } from './modules.utils'
import { Module, Verb, Data } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbId } from './modules.constants'
import { getVerbName } from './modules.utils'
import { ControlBar, CodeBlock } from '../../components'

export const ModulesSchema: React.FC<{
  className: string
  modules: Module[]
  moduleId?: string
  selectedVerbs?: VerbId[]
}> = ({
  className,
  modules,
  moduleId,
  selectedVerbs,
}) => {
  if (!moduleId || !selectedVerbs) return <></>
  const module = modules.find((module) => module?.name === moduleId)

  const verbs: [Verb, Data[]][] = []
  for(const verbId of selectedVerbs) {
    const verb = module?.verbs.find((v) => v.verb?.name === getVerbName(verbId))
    const callData =
    module?.data.filter((data) => [verb?.verb?.request?.name, verb?.verb?.response?.name].includes(data.data?.name)) ??
    []
    verb && verbs.push([verb, callData])
  }
  return (
    <div className={classNames(className, 'flex flex-col')}>
      <ControlBar>
        <ControlBar.Text><span>{moduleId}</span> Selected Verbs Schema(s)</ControlBar.Text>
      </ControlBar>
      <div className='flex-1 overflow-auto'>
        <Tab.Group>
          <Tab.List>
            {
              verbs.map(([verb]) => {
                const name = verb.verb?.name
                return <Tab key={name}>{name}</Tab>
              })
            }
          </Tab.List>
            <Tab.Panels>
            {
              verbs.map(([verb, callData]) => {
                const name = verb.verb?.name
                return (
                  <Tab.Panel key={name}>
                    <CodeBlock
                      key={verb.verb?.name}
                      code={buildVerbSchema(
                        verb?.schema,
                        callData.map((d) => d.schema),
                      )}
                      language='graphql'
                    />
                  </Tab.Panel>
                )
              })
            }
          </Tab.Panels>
        </Tab.Group>
      </div>
    </div>
  )
}