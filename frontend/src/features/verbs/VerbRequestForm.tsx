import { useEffect, useState } from 'react'
import { CodeEditor, type InitialState } from '../../components/CodeEditor'
import { ResizableVerticalPanels } from '../../components/ResizeableVerticalPanels'
import { useClient } from '../../hooks/use-client'
import type { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbService } from '../../protos/xyz/block/ftl/v1/ftl_connect'
import type { Ref } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { classNames } from '../../utils'
import { VerbFormInput } from './VerbFormInput'
import { createVerbRequest, defaultRequest, fullRequestPath, httpPopulatedRequestPath, isHttpIngress, requestType, simpleJsonSchema } from './verb.utils'

export const VerbRequestForm = ({ module, verb }: { module?: Module; verb?: Verb }) => {
  const client = useClient(VerbService)
  const [activeTabId, setActiveTabId] = useState('body')
  const [initialEditorState, setInitialEditorText] = useState<InitialState>({ initialText: '' })
  const [editorText, setEditorText] = useState('')
  const [initialHeadersState, setInitialHeadersText] = useState<InitialState>({ initialText: '' })
  const [headersText, setHeadersText] = useState('')
  const [response, setResponse] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const editorTextKey = `${module?.name}-${verb?.verb?.name}-editor-text`
  const headersTextKey = `${module?.name}-${verb?.verb?.name}-headers-text`

  useEffect(() => {
    if (verb) {
      const savedEditorValue = localStorage.getItem(editorTextKey)
      let editorValue: string
      if (savedEditorValue != null && savedEditorValue !== '') {
        editorValue = savedEditorValue
      } else {
        editorValue = defaultRequest(verb)
      }

      const schemaString = JSON.stringify(simpleJsonSchema(verb))
      setInitialEditorText({ initialText: editorValue, schema: schemaString })
      localStorage.setItem(editorTextKey, editorValue)
      handleEditorTextChanged(editorValue)

      const savedHeadersValue = localStorage.getItem(headersTextKey)
      let headerValue: string
      if (savedHeadersValue != null && savedHeadersValue !== '') {
        headerValue = savedHeadersValue
      } else {
        headerValue = '{\n  "console": ["example"]\n}'
      }
      setInitialHeadersText({ initialText: headerValue })
      setHeadersText(headerValue)
      localStorage.setItem(headersTextKey, headerValue)
    }
  }, [verb, activeTabId])

  const handleEditorTextChanged = (text: string) => {
    setEditorText(text)
    localStorage.setItem(editorTextKey, text)
  }

  const handleHeadersTextChanged = (text: string) => {
    setHeadersText(text)
    localStorage.setItem(headersTextKey, text)
  }

  const handleTabClick = (e: React.MouseEvent<HTMLButtonElement>, id: string) => {
    e.preventDefault()
    setActiveTabId(id)
  }

  const tabs = [{ id: 'body', name: 'Body' }]

  if (isHttpIngress(verb)) {
    tabs.push({ id: 'headers', name: 'Headers' })
  }

  tabs.push({ id: 'verbschema', name: 'Verb Schema' }, { id: 'jsonschema', name: 'JSONSchema' })

  const handleSubmit = async (path: string) => {
    setResponse(null)
    setError(null)

    try {
      const verbRef: Ref = {
        name: verb?.verb?.name,
        module: module?.name,
      } as Ref

      const requestBytes = createVerbRequest(path, verb, editorText, headersText)
      const response = await client.call({ verb: verbRef, body: requestBytes })

      if (response.response.case === 'body') {
        const textDecoder = new TextDecoder('utf-8')
        const jsonString = textDecoder.decode(response.response.value)

        setResponse(JSON.stringify(JSON.parse(jsonString), null, 2))
      } else if (response.response.case === 'error') {
        setError(response.response.value.message)
      }
    } catch (error) {
      console.error('There was an error with the request:', error)
      setError(String(error))
    }
  }

  const bottomText = response ?? error ?? ''

  const bodyEditor = <CodeEditor initialState={initialEditorState} onTextChanged={handleEditorTextChanged} />
  const bodyPanels =
    bottomText === '' ? (
      bodyEditor
    ) : (
      <ResizableVerticalPanels
        topPanelContent={bodyEditor}
        bottomPanelContent={<CodeEditor initialState={{ initialText: bottomText, readonly: true }} onTextChanged={setHeadersText} />}
      />
    )

  return (
    <div className='flex flex-col h-full overflow-hidden pt-4'>
      <VerbFormInput
        requestType={requestType(verb)}
        initialPath={httpPopulatedRequestPath(module, verb)}
        requestPath={fullRequestPath(module, verb)}
        readOnly={!isHttpIngress(verb)}
        onSubmit={handleSubmit}
      />
      <div>
        <div className='border-b border-gray-200 dark:border-white/10'>
          <nav className='-mb-px flex space-x-6 pl-4' aria-label='Tabs'>
            {tabs.map((tab) => (
              <button
                type='button'
                key={tab.name}
                className={classNames(
                  activeTabId === tab.id
                    ? 'border-indigo-500 text-indigo-600 dark:border-indigo-400 dark:text-indigo-400'
                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:hover:border-gray-500 dark:text-gray-500 dark:hover:text-gray-300',
                  'whitespace-nowrap cursor-pointer border-b-2 py-2 px-1 text-sm font-medium',
                )}
                aria-current={activeTabId === tab.id ? 'page' : undefined}
                onClick={(e) => handleTabClick(e, tab.id)}
              >
                {tab.name}
              </button>
            ))}
          </nav>
        </div>
      </div>
      <div className='flex-1 overflow-hidden'>
        <div className='h-full overflow-y-scroll'>
          {activeTabId === 'body' && bodyPanels}
          {activeTabId === 'verbschema' && <CodeEditor initialState={{ initialText: verb?.schema ?? 'what', readonly: true }} />}
          {activeTabId === 'jsonschema' && <CodeEditor initialState={{ initialText: verb?.jsonRequestSchema ?? '', readonly: true }} />}
          {activeTabId === 'headers' && <CodeEditor initialState={initialHeadersState} onTextChanged={handleHeadersTextChanged} />}
        </div>
      </div>
    </div>
  )
}
