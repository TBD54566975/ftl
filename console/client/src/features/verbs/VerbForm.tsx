/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-call */
import {JSONSchemaFaker, Schema} from 'json-schema-faker'
import React, {useEffect} from 'react'
import {CodeBlock} from '../../components/CodeBlock'
import {useClient} from '../../hooks/use-client'
import {Module, Verb} from '../../protos/xyz/block/ftl/v1/console/console_pb'
import {VerbService} from '../../protos/xyz/block/ftl/v1/ftl_connect'
import {VerbRef} from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import Editor from '@monaco-editor/react'
import {useDarkMode} from '../../providers/dark-mode-provider'

type Props = {
  module?: Module
  verb?: Verb
}

export const VerbForm: React.FC<Props> = ({module, verb}) => {
  const client = useClient(VerbService)
  const {isDarkMode} = useDarkMode()
  const [editorText, setEditorText] = React.useState<string>('')
  const [response, setResponse] = React.useState<string | null>(null)
  const [error, setError] = React.useState<string | null>(null)
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  const [schema, setSchema] = React.useState<Schema>()
  useEffect(() => {
    if (verb?.jsonRequestSchema) {
      JSONSchemaFaker.option('maxItems', 2)
      JSONSchemaFaker.option('alwaysFakeOptionals', true)

      // eslint-disable-next-line @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-unsafe-assignment
      const verbSchema = JSON.parse(verb.jsonRequestSchema) as Schema
      setSchema(verbSchema)
      setEditorText(JSON.stringify(JSONSchemaFaker.generate(verbSchema), null, 2))
    }
  }, [])

  function handleEditorChange(value: string | undefined, _) {
    setEditorText(value ?? '')
  }

  // eslint-disable-next-line @typescript-eslint/no-misused-promises
  const handleSubmit: React.FormEventHandler<HTMLFormElement> = async event => {
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
      const response = await client.call({verb: verbRef, body: uint8Array})
      if (response.response.case === 'body') {
        const jsonString = Buffer.from(response.response.value).toString(
          'utf-8'
        )

        setResponse(JSON.stringify(JSON.parse(jsonString), null, 2))
      } else if (response.response.case === 'error') {
        setError(response.response.value.message)
      }
    } catch (error) {
      console.error('There was an error with the request:', error)
      setError(String(error))
    }
  }
  const handleEditorWillMount = monaco => {
    verb?.jsonRequestSchema &&  monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
      validate: true,
      schemas: [{
          uri: 'http://myserver/json-schema',
          fileMatch: ['*'],
          // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
          schema,
      }],
    })
  }

  return (
    <>
      <form
        onSubmit={handleSubmit}
        className='rounded-lg'
      >
        <div className='border border-gray-200 dark:border-slate-800 rounded-sm'>
          <Editor
            height='35vh'
            theme={`${isDarkMode ? 'vs-dark' : 'light'}`}
            defaultLanguage='json'
            defaultValue={editorText}
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
          <CodeBlock
            code={response}
            language='go'
          />
        </div>
      )}
      {error && (
        <div
          className='mt-4 bg-red-100 border-l-4 border-red-500 text-red-700 p-4'
          role='alert'
        >
          {error}
        </div>
      )}
    </>
  )
}
