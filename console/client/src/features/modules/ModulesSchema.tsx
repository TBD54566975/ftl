import React from 'react'
import { Tab } from '@headlessui/react'
import { classNames } from '../../utils'
import { buildVerbSchema } from './modules.utils'
import { Module, Verb, Data } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbId } from './modules.constants'
import { getNames } from './modules.utils'
import { CodeBlock } from '../../components'
import { Panel } from './components'
export const ModulesSchema: React.FC<{
  className: string
  modules: Module[]
  selectedModules?: string[]
  selectedVerbs: VerbId[]
}> = ({ className, modules, selectedVerbs }) => {
  if (!selectedVerbs.length) return <></>
  const verbs: [Verb, Data[]][] = []
  for (const verbId of selectedVerbs) {
    const [moduleName, verbName] = getNames(verbId)
    const module = modules.find((module) => module?.name === moduleName)
    const verb = module?.verbs.find((v) => v.verb?.name === verbName)
    const callData =
      module?.data.filter((data) =>
        [verb?.verb?.request?.name, verb?.verb?.response?.name].includes(data.data?.name),
      ) ?? []
    verb && verbs.push([verb, callData])
  }
  return (
    <Panel className={className}>
      <Panel.Header>Verb Schema(s)</Panel.Header>
      <Panel.Body>
        <Tab.Group>
          <Tab.List>
            {verbs.map(([verb]) => {
              const name = verb.verb?.name
              return <Tab key={name}>{name}</Tab>
            })}
          </Tab.List>
          <Tab.Panels>
            {verbs.map(([verb, callData]) => {
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
            })}
          </Tab.Panels>
        </Tab.Group>
      </Panel.Body>
    </Panel>
  )
}
