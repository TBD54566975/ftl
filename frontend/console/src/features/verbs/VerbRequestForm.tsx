import { Copy01Icon } from 'hugeicons-react'
import { useCallback, useContext, useEffect, useState } from 'react'
import { CodeEditor } from '../../components/CodeEditor'
import { ResizableVerticalPanels } from '../../components/ResizableVerticalPanels'
import { useClient } from '../../hooks/use-client'
import type { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import type { Ref } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { VerbService } from '../../protos/xyz/block/ftl/v1/verb_connect'
import { NotificationType, NotificationsContext } from '../../providers/notifications-provider'
import { classNames } from '../../utils'
import { VerbFormInput } from './VerbFormInput'
import {
  createVerbRequest as createCallRequest,
  defaultRequest,
  fullRequestPath,
  generateCliCommand,
  httpPopulatedRequestPath,
  isHttpIngress,
  requestType,
  simpleJsonSchema,
} from './verb.utils'

export const VerbRequestForm = ({ module, verb }: { module?: Module; verb?: Verb }) => {
  const client = useClient(VerbService)
  const { showNotification } = useContext(NotificationsContext)
  const [activeTabId, setActiveTabId] = useState('body')
  const [bodyText, setBodyText] = useState('')
  const [headersText, setHeadersText] = useState('')
  const [response, setResponse] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [path, setPath] = useState('')

  const bodyTextKey = `${module?.name}-${verb?.verb?.name}-body-text`
  const headersTextKey = `${module?.name}-${verb?.verb?.name}-headers-text`

  useEffect(() => {
    setPath(httpPopulatedRequestPath(module, verb))
  }, [module, verb])

  useEffect(() => {
    if (verb) {
      const savedBodyValue = localStorage.getItem(bodyTextKey)
      const bodyValue = savedBodyValue ?? defaultRequest(verb)
      setBodyText(bodyValue)

      const savedHeadersValue = localStorage.getItem(headersTextKey)
      const headerValue = savedHeadersValue ?? '{}'
      setHeadersText(headerValue)

      setResponse(null)
      setError(null)
    }
  }, [verb, activeTabId])

  const handleBodyTextChanged = (text: string) => {
    setBodyText(text)
    localStorage.setItem(bodyTextKey, text)
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

  const httpCall = (path: string) => {
    const method = requestType(verb)

    fetch(path, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...JSON.parse(headersText),
      },
      ...(method === 'POST' || method === 'PUT' ? { body: bodyText } : {}),
    })
      .then(async (response) => {
        if (response.ok) {
          const json = await response.json()
          setResponse(JSON.stringify(json, null, 2))
        } else {
          const text = await response.text()
          setError(text)
        }
      })
      .catch((error) => {
        setError(String(error))
      })
  }

  const ftlCall = (path: string) => {
    const verbRef: Ref = {
      name: verb?.verb?.name,
      module: module?.name,
    } as Ref

    const requestBytes = createCallRequest(path, verb, bodyText, headersText)
    client
      .call({ verb: verbRef, body: requestBytes })
      .then((response) => {
        if (response.response.case === 'body') {
          const textDecoder = new TextDecoder('utf-8')
          const jsonString = textDecoder.decode(response.response.value)

          setResponse(JSON.stringify(JSON.parse(jsonString), null, 2))
        } else if (response.response.case === 'error') {
          setError(response.response.value.message)
        }
      })
      .catch((error) => {
        console.error(error)
      })
  }

  const handleSubmit = async (path: string) => {
    setResponse('')
    setError('')

    try {
      if (isHttpIngress(verb)) {
        httpCall(path)
      } else {
        ftlCall(path)
      }
    } catch (error) {
      console.error('There was an error with the request:', error)
      setError(String(error))
    }
  }

  const handleCopyButton = () => {
    if (!verb) {
      return
    }

    const cliCommand = generateCliCommand(verb, path, headersText, bodyText)
    navigator.clipboard
      .writeText(cliCommand)
      .then(() => {
        showNotification({
          title: 'Copied to clipboard',
          message: cliCommand,
          type: NotificationType.Info,
        })
      })
      .catch((err) => {
        console.error('Failed to copy text: ', err)
      })
  }

  const handleResetBody = useCallback(() => {
    if (verb) {
      handleBodyTextChanged(defaultRequest(verb))
    }
  }, [verb, bodyTextKey])

  const bottomText = response || error || ''
  const schemaString = verb ? JSON.stringify(simpleJsonSchema(verb)) : ''

  return (
    <div className='flex flex-col h-full overflow-hidden pt-4'>
      <VerbFormInput
        requestType={requestType(verb)}
        path={path}
        setPath={setPath}
        requestPath={fullRequestPath(module, verb)}
        readOnly={!isHttpIngress(verb)}
        onSubmit={handleSubmit}
      />
      <div>
        <div className='border-b border-gray-200 dark:border-white/10'>
          <div className='flex justify-between items-center  pr-4'>
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
            <button
              type='button'
              title='Copy'
              className='flex items-center p-1 rounded text-indigo-500 hover:bg-gray-200 dark:hover:bg-gray-600 cursor-pointer'
              onClick={handleCopyButton}
            >
              <Copy01Icon className='size-5' />
            </button>
          </div>
        </div>
      </div>
      <div className='flex-1 overflow-hidden'>
        <div className='h-full overflow-y-scroll'>
          {activeTabId === 'body' && (
            <ResizableVerticalPanels
              topPanelContent={
                <div className='relative h-full'>
                  <button
                    type='button'
                    onClick={handleResetBody}
                    className='text-sm absolute top-2 right-2 z-10 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 py-1 px-2 rounded'
                  >
                    Reset
                  </button>
                  <CodeEditor id='body-editor' value={bodyText} onTextChanged={handleBodyTextChanged} schema={schemaString} />
                </div>
              }
              bottomPanelContent={bottomText !== '' ? <CodeEditor id='response-editor' value={bottomText} readonly /> : null}
            />
          )}
          {activeTabId === 'verbschema' && <CodeEditor readonly value={verb?.schema ?? ''} />}
          {activeTabId === 'jsonschema' && <CodeEditor readonly value={verb?.jsonRequestSchema ?? ''} />}
          {activeTabId === 'headers' && <CodeEditor value={headersText} onTextChanged={handleHeadersTextChanged} />}
        </div>
      </div>
    </div>
  )
}
