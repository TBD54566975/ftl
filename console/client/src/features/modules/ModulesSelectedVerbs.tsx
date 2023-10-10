import { Tab } from '@headlessui/react'
import React from 'react'
import { CodeBlock } from '../../components'
import { Data, Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { classNames } from '../../utils'
import { VerbForm } from '../verbs/VerbForm'
import { Panel } from './components'
import { VerbId } from './modules.constants'
import { buildVerbSchema, getNames } from './modules.utils'

export const ModulesSelectedVerbs = ({
  className,
  modules,
  selectedVerbs,
}: {
  className?: string
  modules: Module[]
  selectedVerbs?: VerbId[]
}) => {
  if (!selectedVerbs?.length) return <></>
  const verbs: { module: Module; verb: Verb; callData: Data[] }[] = []
  for (const verbId of selectedVerbs) {
    const [moduleName, verbName] = getNames(verbId)
    const module = modules.find((module) => module?.name === moduleName)
    const verb = module?.verbs.find((v) => v.verb?.name === verbName)
    const callData =
      module?.data.filter((data) =>
        [verb?.verb?.request?.name, verb?.verb?.response?.name].includes(data.data?.name),
      ) ?? []
    if (verb && module) verbs.push({ module, verb, callData })
  }
  return (
    <Panel className={className}>
      <Tab.Group>
        <Panel.Header className='shadow dark:shadow'>
          <Tab.List className='h-full flex gap-x-0.5'>
            {verbs.map(({ verb, module }) => {
              const name = verb.verb?.name
              const id = `${module.name}.${name}`
              return (
                <Tab key={id} as={React.Fragment}>
                  {({ selected }) => (
                    <button
                      className={classNames(
                        `rounded-t-md px-4 py-2`,
                        selected ? 'dark:bg-gray-800 bg-white' : 'bg-gray-200 dark:bg-gray-700',
                      )}
                    >
                      {id}
                    </button>
                  )}
                </Tab>
              )
            })}
          </Tab.List>
        </Panel.Header>
        <Panel.Body className={`flex flex-col gap-4 dark:bg-gray-800 bg-white p-2`}>
          {verbs.map(({ module, verb, callData }) => (
            <Tab.Panel key={verb.verb?.name} className={`flex flex-col gap-4`}>
              <CodeBlock
                key={verb.verb?.name}
                code={buildVerbSchema(
                  verb?.schema,
                  callData.map((d) => d.schema),
                )}
                language='graphql'
              />
              <VerbForm module={module} verb={verb} />
            </Tab.Panel>
          ))}
        </Panel.Body>
      </Tab.Group>
    </Panel>
  )
}
