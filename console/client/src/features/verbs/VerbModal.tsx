import React from 'react'
import {   useSearchParams, useNavigate , useLocation } from 'react-router-dom'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { modulesContext } from '../../providers/modules-provider.tsx'
import { getCodeBlock } from '../../utils/data.utils.ts'
import { classNames } from '../../utils/react.utils.ts'
import { getCalls, getVerbCode } from './verb.utils.ts'
import { VerbCalls } from './VerbCalls.tsx'
import { Dialog, Transition } from '@headlessui/react'
import { ChevronRightIcon } from '@heroicons/react/20/solid'
export function VerbModal() {
  const [ searchParams ] = useSearchParams()
  const verbName= searchParams.get('verb')
  const moduleName = searchParams.get('module') 
  const modules = React.useContext(modulesContext)
  
  const module = modules.modules.find(m => m.name === moduleName)
  const verb = module?.verbs.find(v => v.verb?.name === verbName?.toLocaleLowerCase())

  const callData = module?.data.filter(data =>
    [ verb?.verb?.request?.name, verb?.verb?.response?.name ].includes(data.name)
  )

  const navigate = useNavigate()
  const location = useLocation()

  
  const handleClose = () =>{
    searchParams.delete('verb')
    navigate({ ...location, search: searchParams.toString() })
  }

  return (
    <Transition appear
      show={!!verbName}
      as={React.Fragment}
    >
      <Dialog 
        onClose={handleClose}
        as='div'
        className='relative z-10'
      >
        <Transition.Child
          as={React.Fragment}
          enter='ease-out duration-300'
          enterFrom='opacity-0'
          enterTo='opacity-100'
          leave='ease-in duration-200'
          leaveFrom='opacity-100'
          leaveTo='opacity-0'
        >
          <div className='fixed inset-0 bg-black bg-opacity-25' />
        </Transition.Child>
        <div className='fixed inset-0 overflow-y-auto'>
          <div className='flex min-h-full items-center justify-center p-4 text-center'>
            <Transition.Child
              as={React.Fragment}
              enter='ease-out duration-300'
              enterFrom='opacity-0 scale-95'
              enterTo='opacity-100 scale-100'
              leave='ease-in duration-200'
              leaveFrom='opacity-100 scale-100'
              leaveTo='opacity-0 scale-95'
            >
              <Dialog.Panel className={`w-full max-w-4xl transform overflow-hidden rounded-2xl bg-white p-6 text-left align-middle shadow-xl transition-all`}>
                <Dialog.Title
                  as='h3'
                  className='text-lg font-medium leading-6 text-gray-900'
                >
                  <ol role='list'
                    className='flex items-center space-x-4'
                  >
                    <li>
                      <div className='flex items-center'>
                        <button className='focus:outline-none'
                          onClick={handleClose}
                        >
                          <span className='capitalize ml-4 text-sm font-medium text-gray-400 hover:text-gray-500'>{moduleName} (module)</span>
                        </button>
                      </div>
                    </li>
                    <li>
                      <div className='flex items-center'>
                        <ChevronRightIcon className='h-5 w-5 flex-shrink-0 text-gray-400'
                          aria-hidden='true'
                        />
                        <span className='capitalize ml-4 text-sm font-medium text-gray-400 hover:text-gray-500'>{verbName} (verb)</span>
                      </div>
                    </li>
                  </ol>
                 
                </Dialog.Title>
                <div className='min-w-0 flex-auto'>
                  <div className='text-sm pt-4'>
                    <SyntaxHighlighter language={module?.language || 'go'}
                      style={atomDark}
                    >
                      {getVerbCode(verb?.verb)}
                    </SyntaxHighlighter>
                  </div>
                  <div className='pt-4'>
                    {callData?.map(data => (
                      <div key={data.name}
                        className='text-sm'
                      >
                        <SyntaxHighlighter language='go'
                          style={atomDark}
                        >
                          {getCodeBlock(data)}
                        </SyntaxHighlighter>
                      </div>
                    ))}
                  </div>
                  <div className='flex items-center gap-x-3 pt-6'>
                    <h2 className='min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white'>
                      <div className='flex gap-x-2'>
                        <span className='truncate'>Calls</span>
                      </div>
                    </h2>
                  </div>
                  {getCalls(verb?.verb).map(call =>
                    call.calls.map(call => (
                      <button key={`/modules/${call.module}/verbs/${call.name}`}
                        onClick={ () => {
                          searchParams.set('module', call.module)
                          searchParams.set('verb', call.name)
                          navigate({ ...location, search: searchParams.toString() })
                        }}
                      >
                        <span
                          className={classNames(
                            'text-indigo-400 bg-indigo-400/10 ring-indigo-400/30',
                            'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
                          )}
                        >
                          {call.name}
                        </span>
                      </button>
                    ))
                  )}

                  <VerbCalls module={module}
                    verb={verb}
                  />

                  <div className='flex items-center gap-x-3 pt-6'>
                    <h2 className='min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white'>
                      <div className='flex gap-x-2'>
                        <span className='truncate'>Errors</span>
                      </div>
                    </h2>
                  </div>
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition>
  )
}
