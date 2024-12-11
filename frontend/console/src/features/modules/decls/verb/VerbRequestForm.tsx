import { useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { Button } from '../../../../components/Button'
import { CodeEditor } from '../../../../components/CodeEditor'
import { ResizableVerticalPanels } from '../../../../components/ResizableVerticalPanels'
import { useClient } from '../../../../hooks/use-client'
import { ConsoleService } from '../../../../protos/xyz/block/ftl/console/v1/console_connect'
import type { Module, Verb } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { Ref } from '../../../../protos/xyz/block/ftl/schema/v1/schema_pb'
import { NotificationType, NotificationsContext } from '../../../../providers/notifications-provider'
import { classNames } from '../../../../utils'
import { KeyValuePairForm } from '../KeyValuePairForm'
import { useKeyValuePairs } from '../hooks/useKeyValuePairs'
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
  const client = useClient(ConsoleService)
  const { showNotification } = useContext(NotificationsContext)
  const [activeTabId, setActiveTabId] = useState('body')
  const [bodyText, setBodyText] = useState('')
  const [response, setResponse] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [path, setPath] = useState('')

  const bodyTextKey = `${module?.name}-${verb?.verb?.name}-body-text`
  const headersTextKey = `${module?.name}-${verb?.verb?.name}-headers-text`
  const queryParamsTextKey = `${module?.name}-${verb?.verb?.name}-query-params-text`
  const methodsWithBody = ['POST', 'PUT', 'PATCH', 'CALL', 'CRON', 'SUB']

  const tabs = useMemo(() => {
    const method = requestType(verb)
    const tabsArray = methodsWithBody.includes(method) ? [{ id: 'body', name: 'Body' }] : []

    if (isHttpIngress(verb)) {
      tabsArray.push({ id: 'headers', name: 'Headers' })

      if (['GET', 'DELETE', 'HEAD', 'OPTIONS'].includes(method)) {
        tabsArray.push({ id: 'queryParams', name: 'Query Params' })
      }
    }

    return tabsArray
  }, [module, verb])

  const { pairs: headers, updatePairs: handleHeadersChanged, getPairsObject: getHeadersObject } = useKeyValuePairs(headersTextKey)

  const { pairs: queryParams, updatePairs: handleQueryParamsChanged, getPairsObject: getQueryParamsObject } = useKeyValuePairs(queryParamsTextKey)

  useEffect(() => {
    if (verb) {
      const savedBodyValue = localStorage.getItem(bodyTextKey)
      const bodyValue = savedBodyValue ?? defaultRequest(verb)
      setBodyText(bodyValue)

      setResponse(null)
      setError(null)

      // Set initial tab only when verb changes
      setActiveTabId(tabs[0].id)
    }
  }, [verb, bodyTextKey, tabs])

  useEffect(() => {
    setPath(httpPopulatedRequestPath(module, verb))
  }, [module, verb])

  const handleBodyTextChanged = (text: string) => {
    setBodyText(text)
    localStorage.setItem(bodyTextKey, text)
  }

  const handleTabClick = (e: React.MouseEvent<HTMLButtonElement>, id: string) => {
    e.preventDefault()
    setActiveTabId(id)
  }

  const httpCall = (path: string) => {
    const method = requestType(verb)
    const headerObject = getHeadersObject()
    const queryParamsObject = getQueryParamsObject()

    // Construct URL with query parameters
    const url = new URL(path)
    for (const [key, value] of Object.entries(queryParamsObject)) {
      url.searchParams.append(key, value)
    }

    fetch(url.toString(), {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...headerObject,
      },
      ...(methodsWithBody.includes(method) ? { body: bodyText } : {}),
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

    const requestBytes = createCallRequest(path, verb, bodyText, JSON.stringify(headers))
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

    const cliCommand = generateCliCommand(verb, path, JSON.stringify(headers), bodyText)
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

  const handleCopyBody = () => {
    navigator.clipboard
      .writeText(bodyText)
      .then(() => {
        showNotification({
          title: 'Copied to clipboard',
          message: 'Request body copied to clipboard',
          type: NotificationType.Info,
        })
      })
      .catch((err) => {
        console.error('Failed to copy text: ', err)
      })
  }

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
        handleCopyButton={handleCopyButton}
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
          </div>
        </div>
      </div>
      <div className='flex-1 overflow-hidden flex flex-col'>
        <ResizableVerticalPanels
          topPanelContent={
            <div className='h-full overflow-auto'>
              {activeTabId === 'body' && (
                <div className='h-full'>
                  <div className='relative h-full'>
                    <div className='absolute top-2 right-2 z-10 flex gap-2'>
                      <Button variant='secondary' size='xs' title='Copy' onClick={handleCopyBody}>
                        Copy
                      </Button>
                      <Button variant='secondary' size='xs' type='button' title='Reset' onClick={handleResetBody}>
                        Reset
                      </Button>
                    </div>
                    <CodeEditor id='body-editor' value={bodyText} onTextChanged={handleBodyTextChanged} schema={schemaString} />
                  </div>
                </div>
              )}

              {activeTabId === 'headers' && <KeyValuePairForm keyValuePairs={headers} onChange={handleHeadersChanged} />}
              {activeTabId === 'queryParams' && <KeyValuePairForm keyValuePairs={queryParams} onChange={handleQueryParamsChanged} />}
            </div>
          }
          bottomPanelContent={bottomText !== '' ? <CodeEditor id='response-editor' value={bottomText} readonly /> : null}
        />
      </div>
    </div>
  )
}
