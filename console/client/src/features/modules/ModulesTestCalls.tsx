import React from 'react'
import { Tab } from '@headlessui/react'
import Editor, { Monaco } from '@monaco-editor/react'
import type { JSONSchema4, JSONSchema6, JSONSchema7 } from 'json-schema'
import { JSONSchemaFaker } from 'json-schema-faker'
import { CodeBlock } from '../../components'
import { Panel } from './components'
import { useClient } from '../../hooks/use-client'
import { VerbService } from '../../protos/xyz/block/ftl/v1/ftl_connect'
import { VerbRef } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { useDarkMode } from '../../providers/dark-mode-provider'
import { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbId } from './modules.constants'
import { getNames } from './modules.utils'

export type Schema = JSONSchema4 | JSONSchema6 | JSONSchema7

interface Props {
  module?: Module
  verb?: Verb
}

const VerbForm = ({ module, verb }: Props) => {
  const client = useClient(VerbService)
  const { isDarkMode } = useDarkMode()
  const [editorText, setEditorText] = React.useState<string>('')
  const [response, setResponse] = React.useState<string | null>(null)
  const [error, setError] = React.useState<string | null>(null)
  const [schema, setSchema] = React.useState<Schema>()
  const [monaco, setMonaco] = React.useState<Monaco>()

  React.useEffect(() => {
    if (verb?.jsonRequestSchema) {
      JSONSchemaFaker.option('maxItems', 2)
      JSONSchemaFaker.option('alwaysFakeOptionals', true)

      // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
      const verbSchema = JSON.parse(verb.jsonRequestSchema) as Schema
      setSchema(verbSchema)
      setEditorText(JSON.stringify(JSONSchemaFaker.generate(verbSchema), null, 2))
    }
  }, [module, verb])

  const handleEditorChange = (value: string | undefined) => {
    setEditorText(value ?? '')
  }

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = async (event) => {
    event.preventDefault()

    setResponse(null)
    setError(null)

    try {
      const verbRef: VerbRef = {
        name: verb?.verb?.name,
        module: module?.name,
      } as VerbRef

      const buffer = Buffer.from(editorText)
      const uint8Array = new Uint8Array(buffer)
      const response = await client.call({ verb: verbRef, body: uint8Array })
      if (response.response.case === 'body') {
        const jsonString = Buffer.from(response.response.value).toString('utf-8')

        setResponse(JSON.stringify(JSON.parse(jsonString), null, 2))
      } else if (response.response.case === 'error') {
        setError(response.response.value.message)
      }
    } catch (error) {
      console.error('There was an error with the request:', error)
      setError(String(error))
    }
  }
  const handleEditorWillMount = (monaco: Monaco) => {
    setMonaco(monaco)
  }

  React.useEffect(() => {
    schema &&
      monaco?.languages.json.jsonDefaults.setDiagnosticsOptions({
        validate: true,
        schemas: [{ schema, uri: 'http://myserver/foo-schema.json', fileMatch: ['*'] }],
      })
  }, [monaco, schema])

  return (
    <>
      <form onSubmit={handleSubmit} className='rounded-lg'>
        <div className='border border-gray-200 dark:border-slate-800 rounded-sm'>
          <Editor
            key={[module?.name, verb?.verb?.name].join('.')}
            height='35vh'
            theme={`${isDarkMode ? 'vs-dark' : 'light'}`}
            defaultLanguage='json'
            path={[module?.name, verb?.verb?.name].join('.')}
            value={editorText}
            options={{
              lineNumbers: 'off',
              scrollBeyondLastLine: false,
            }}
            onChange={handleEditorChange}
            beforeMount={handleEditorWillMount}
          />
        </div>

        <button
          type='submit'
          className='bg-indigo-700 text-white mt-4 px-4 py-2 rounded hover:bg-indigo-600 focus:outline-none focus:bg-indigo-600'
        >
          Submit
        </button>
      </form>
      {response && (
        <div className='pt-4'>
          <CodeBlock code={response} language='go' />
        </div>
      )}
      {error && (
        <div className='mt-4 bg-red-100 border-l-4 border-red-500 text-red-700 p-4' role='alert'>
          {error}
        </div>
      )}
    </>
  )
}

export const ModulesTestCalls: React.FC<{
  className: string
  modules: Module[]
  selectedVerbs?: VerbId[]
}> = ({ className, modules, selectedVerbs }) => {
  if (!selectedVerbs?.length) return <></>
  const verbs: [Module, Verb][] = []
  for (const verbId of selectedVerbs) {
    const [moduleName, verbName] = getNames(verbId)
    const module = modules.find((module) => module?.name === moduleName)
    const verb = module?.verbs.find((v) => v.verb?.name === verbName)
    if (verb && module) verbs.push([module, verb])
  }
  return (
    <Panel className={className}>
      <Panel.Header>Verb Test Call(s)</Panel.Header>
      <Panel.Body>
        <Tab.Group>
          <Tab.List>
            {verbs.map(([_, verb]) => {
              const name = verb.verb?.name
              return <Tab key={name}>{name}</Tab>
            })}
          </Tab.List>
          {verbs.map(([module, verb]) => (
            <Tab.Panel key={verb.verb?.name}>
              <VerbForm module={module} verb={verb} />
            </Tab.Panel>
          ))}
        </Tab.Group>
      </Panel.Body>
    </Panel>
  )
}
